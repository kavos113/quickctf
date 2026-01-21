package manifest

import "github.com/opencontainers/go-digest"

type Descriptor struct {
	MediaType    string             `json:"mediaType"`
	Digest       digest.Digest      `json:"digest"`
	Size         int64              `json:"size"`
	Urls         *[]string          `json:"urls,omitempty"`
	Annotations  *map[string]string `json:"annotations,omitempty"`
	Data         *string            `json:"data,omitempty"`
	ArtifactType *string            `json:"artifactType,omitempty"`
}

type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	ArtifactType  *string `json:"artifactType,omitempty"`

	// for Image Manifest
	Config *Descriptor   `json:"config,omitempty"`
	Layers *[]Descriptor `json:"layers,omitempty"`

	// for Image Index
	Manifests *[]Descriptor `json:"manifests,omitempty"`

	Subject     *Descriptor        `json:"subject,omitempty"`
	Annotations *map[string]string `json:"annotations,omitempty"`
}
