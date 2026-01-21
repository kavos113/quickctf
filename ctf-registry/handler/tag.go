package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/kavos113/quickctf/ctf-registry/store"
	"github.com/labstack/echo/v4"
)

type TagHandler struct {
	ms store.Store
}

func NewTagHandler(s store.Store) *TagHandler {
	return &TagHandler{ms: s}
}

type tagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (h *TagHandler) GetTags(c echo.Context) error {
	name := c.Param("name")
	last := c.QueryParam("last")
	nstr := c.QueryParam("n")

	n := -1
	if nstr != "" {
		ni, err := strconv.Atoi(nstr)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		n = ni
	}

	tags, err := h.ms.GetTagList(name, n, last)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	res := &tagsResponse{
		Name: name,
		Tags: tags,
	}
	return c.JSON(http.StatusOK, res)
}

func (h *TagHandler) DeleteTag(c echo.Context) error {
	name := c.Param("name")
	tag := c.Param("tag")

	if err := h.ms.DeleteTag(name, tag); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusAccepted)
}
