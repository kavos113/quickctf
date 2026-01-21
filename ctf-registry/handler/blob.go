package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobHandler struct {
	bs storage.Storage
}

func NewBlobHandler(s storage.Storage) *BlobHandler {
	return &BlobHandler{bs: s}
}

func (h *BlobHandler) GetBlobs(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	size, err := h.bs.ReadBlobToWriter(name, d, c.Response().Writer)
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

	err = h.bs.DeleteBlob(name, d)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
}
