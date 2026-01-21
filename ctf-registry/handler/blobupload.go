package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobUploadHandler struct {
	bs storage.Storage
}

func NewBlobUploadHandler(s storage.Storage) *BlobUploadHandler {
	return &BlobUploadHandler{bs: s}
}

func (h *BlobUploadHandler) GetBlobUploads(c echo.Context) error {
	name := c.Param("name")
	ref := c.Param("reference")

	size, err := h.bs.GetUploadBlobSize(ref)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("Range", fmt.Sprintf("0-%d", size-1))
	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, ref))
	return c.NoContent(http.StatusNoContent)
}

func (h *BlobUploadHandler) PostBlobUploads(c echo.Context) error {
	name := c.Param("name")

	id, err := uuid.NewV7()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to generate upload ID")
	}
	c.Response().Header().Set("Docker-Upload-UUID", id.String())

	dstr := c.QueryParam("digest")
	if dstr != "" {
		// monolithic upload
		d, err := digest.Parse(dstr)
		if err != nil {
			log.Printf("cannot parse digest %s: %+v", dstr, err)
			return c.String(http.StatusBadRequest, "invalid digest format")
		}

		_, err = h.bs.UploadBlob(id.String(), c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		err = h.bs.CommitBlob(name, id.String(), d)
		if err != nil {
			if errors.Is(err, storage.ErrNotVerified) {
				return c.NoContent(http.StatusBadRequest)
			}
			return c.NoContent(http.StatusInternalServerError)
		}

		c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/%s", name, d.String()))
		return c.NoContent(http.StatusCreated)
	}

	mount := c.QueryParam("mount")
	from := c.QueryParam("from")
	if mount != "" && from != "" {
		// mount from another repository
		md, err := digest.Parse(mount)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		if err := h.bs.LinkBlob(name, from, md); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, id.String()))
				return c.NoContent(http.StatusAccepted)
			}
			return c.NoContent(http.StatusInternalServerError)
		}

		c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/%s", name, md.String()))
		return c.NoContent(http.StatusCreated)
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, id.String()))

	return c.NoContent(http.StatusAccepted)
}

func (h *BlobUploadHandler) PatchBlobUpload(c echo.Context) error {
	name := c.Param("name")
	reference := c.Param("reference")

	cr := c.Request().Header.Get("Content-Range")
	cl := c.Request().Header.Get("Content-Length")
	if cr != "" && cl != "" {
		s, e, err := parseContentRange(cr)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if s > e {
			return c.NoContent(http.StatusRequestedRangeNotSatisfiable)
		}

		length, err := strconv.ParseInt(cl, 10, 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		if length != (e-s)+1 {
			return c.NoContent(http.StatusRequestedRangeNotSatisfiable)
		}

		size, err := h.bs.GetUploadBlobSize(reference)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusInternalServerError)
		}
		if size != s {
			return c.NoContent(http.StatusRequestedRangeNotSatisfiable)
		}
	}

	size, err := h.bs.UploadBlob(reference, c.Request().Body)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidRange) {
			return c.NoContent(http.StatusRequestedRangeNotSatisfiable)
		}
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, reference))
	c.Response().Header().Set("Range", fmt.Sprintf("0-%d", size-1))

	return c.NoContent(http.StatusAccepted)
}

func (h *BlobUploadHandler) PutBlobUpload(c echo.Context) error {
	name := c.Param("name")
	reference := c.Param("reference")

	dstr := c.QueryParam("digest")
	if dstr == "" {
		return c.String(http.StatusBadRequest, "digest query parameter is required")
	}

	d, err := digest.Parse(dstr)
	if err != nil {
		log.Printf("cannot parse digest %s: %+v", dstr, err)
		return c.String(http.StatusBadRequest, "invalid digest format")
	}

	_, err = h.bs.UploadBlob(reference, c.Request().Body)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	err = h.bs.CommitBlob(name, reference, d)
	if err != nil {
		if errors.Is(err, storage.ErrNotVerified) {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/%s", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}

func parseContentRange(r string) (int64, int64, error) {
	s, e, ok := strings.Cut(r, "-")
	if !ok {
		return 0, 0, errors.New("no separator")
	}
	start, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, 0, err
	}
	end, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return start, end, err
}
