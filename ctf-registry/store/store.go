package store

import (
	"github.com/kavos113/quickctf/ctf-registry/manifest"
	"github.com/opencontainers/go-digest"
)

type Store interface {
	SaveTag(repoName string, d digest.Digest, tag string) error
	ReadTag(repoName string, tag string) (string, error)
	DeleteTag(repoName string, tag string) error
	// GetTagList limit(default: -1), last: optional
	GetTagList(repoName string, limit int, last string) ([]string, error)

	// AddReference d: parent digest, desc: child descriptor
	AddReference(repoName string, d digest.Digest, desc manifest.Descriptor) error
	GetReferences(repoName string, d digest.Digest, artifactType string) ([]manifest.Descriptor, error)

	// Blob-Repository association
	AddBlob(repoName string, d digest.Digest) error
	DeleteBlob(repoName string, d digest.Digest) error
	IsExistBlob(repoName string, d digest.Digest) (bool, error)
	// LinkBlob associates an existing blob with a new repository (for cross-repo mount)
	LinkBlob(newRepo string, d digest.Digest) error
}
