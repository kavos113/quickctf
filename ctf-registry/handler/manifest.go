package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kavos113/quickctf/ctf-registry/manifest"
	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/kavos113/quickctf/ctf-registry/store"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type ManifestHandler struct {
	bs storage.Storage
	ms store.Store
}

func NewManifestHandler(bs storage.Storage, ms store.Store) *ManifestHandler {
	return &ManifestHandler{
		bs: bs,
		ms: ms,
	}
}

const (
	mediaTypeOCIImageIndex = "application/vnd.oci.image.index.v1+json"
)

func (h *ManifestHandler) PutManifests(c echo.Context) error {
	name := c.Param("name")
	ref := c.Param("reference")
	istag := isTag(ref)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to read manifest")
	}

	var m manifest.Manifest
	if err := json.Unmarshal(payload, &m); err != nil {
		return c.String(http.StatusBadRequest, "invalid manifest")
	}

	if m.Layers != nil && m.Config != nil {
		for _, desc := range append(*m.Layers, *m.Config) {
			err = desc.Digest.Validate()
			if err != nil {
				return c.String(http.StatusBadRequest, "invalid digest")
			}

			exist, err := h.ms.IsExistBlob(name, desc.Digest)
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			if !exist {
				return c.String(http.StatusBadRequest, "unknown blob layer")
			}
		}
	}

	d := digest.FromBytes(payload)

	if err := h.bs.SaveBlob(d, payload); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	// Associate manifest blob with repository
	if err := h.ms.AddBlob(name, d); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	if istag {
		if err := h.ms.SaveTag(name, d, ref); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	if m.Subject != nil {
		artifact := ""
		if m.ArtifactType != nil {
			artifact = *m.ArtifactType
		} else if m.Config != nil {
			artifact = m.Config.MediaType
		}

		desc := manifest.Descriptor{
			MediaType:    m.MediaType,
			Digest:       d,
			Size:         int64(len(payload)),
			Annotations:  m.Annotations,
			ArtifactType: &artifact,
		}

		if err := h.ms.AddReference(name, m.Subject.Digest, desc); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		c.Response().Header().Set("OCI-Subject", m.Subject.Digest.String())
	}

	c.Response().Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s/", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}

func (h *ManifestHandler) GetManifests(c echo.Context) error {
	name := c.Param("name")
	ref := c.Param("reference")
	istag := isTag(ref)

	dstr := ref
	if istag {
		tag, err := h.ms.ReadTag(name, ref)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return c.NoContent(http.StatusNotFound)
			}
			return c.NoContent(http.StatusInternalServerError)
		}
		dstr = tag
	}

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	// Check if manifest is associated with this repository
	exists, err := h.ms.IsExistBlob(name, d)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	if !exists {
		return c.NoContent(http.StatusNotFound)
	}

	rawManifest, err := h.bs.ReadBlob(d)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	var m manifest.Manifest
	if err = json.Unmarshal(rawManifest, &m); err != nil {
		return c.String(http.StatusInternalServerError, "failed to parse json manifest")
	}

	c.Response().Header().Set(echo.HeaderContentType, m.MediaType)
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.JSON(http.StatusOK, m)
}

func (h *ManifestHandler) DeleteManifests(c echo.Context) error {
	name := c.Param("name")
	ref := c.Param("digest")

	istag := isTag(ref)
	if istag {
		err := h.ms.DeleteTag(name, ref)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return c.NoContent(http.StatusNotFound)
			}
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusAccepted)
	}

	d, err := digest.Parse(ref)
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

	return c.NoContent(http.StatusAccepted)
}

func isTag(reference string) bool {
	_, err := digest.Parse(reference)
	return err != nil
}

func (h *ManifestHandler) GetReferrers(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")
	artifact := c.QueryParam("artifactType")

	c.Response().Header().Set(echo.HeaderContentType, mediaTypeOCIImageIndex)
	if artifact != "" {
		c.Response().Header().Set("OCI-Filters-Applied", "artifactType")
	}

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	descs, err := h.ms.GetReferences(name, d, artifact)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			response := manifest.Manifest{
				SchemaVersion: 2,
				MediaType:     mediaTypeOCIImageIndex,
				Manifests:     nil,
			}
			return c.JSON(http.StatusOK, response)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	response := manifest.Manifest{
		SchemaVersion: 2,
		MediaType:     mediaTypeOCIImageIndex,
		Manifests:     &descs,
	}

	return c.JSON(http.StatusOK, response)
}
