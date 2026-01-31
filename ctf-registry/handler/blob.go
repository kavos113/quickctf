package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/kavos113/quickctf/ctf-registry/store"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobHandler struct {
	bs storage.Storage
	ms store.Store
}

func NewBlobHandler(s storage.Storage, ms store.Store) *BlobHandler {
	return &BlobHandler{bs: s, ms: ms}
}

func (h *BlobHandler) GetBlobs(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Check if blob is associated with this repository
	exists, err := h.ms.IsExistBlob(name, d)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	if !exists {
		return c.NoContent(http.StatusNotFound)
	}

	size, err := h.bs.ReadBlobToWriter(d, c.Response().Writer)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		if errors.Is(err, storage.ErrNotVerified) {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("Docker-Content-Digest", d.String())
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(size, 10))

	return c.NoContent(http.StatusOK)
}

func (h *BlobHandler) DeleteBlob(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Delete association from store
	err = h.ms.DeleteBlob(name, d)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	// Note: We don't delete the actual blob from storage here
	// as it may be referenced by other repositories.
	// A garbage collection process should handle orphaned blobs.

	return c.NoContent(http.StatusAccepted)
}
