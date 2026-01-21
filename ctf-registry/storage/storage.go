package storage

import (
	"io"

	"github.com/opencontainers/go-digest"
)

type Storage interface {
	// GetUploadBlobSize return 0 if not found
	GetUploadBlobSize(id string) (int64, error)

	// UploadBlob return blob size
	UploadBlob(id string, r io.Reader) (int64, error)
	CommitBlob(repoName string, id string, d digest.Digest) error
	SaveBlob(repoName string, d digest.Digest, data []byte) error
	ReadBlob(repoName string, d digest.Digest) ([]byte, error)
	// ReadBlobToWriter write blob to w, and verify digest
	ReadBlobToWriter(repoName string, d digest.Digest, w io.Writer) (int64, error)
	IsExistBlob(repoName string, d digest.Digest) (bool, error)
	DeleteBlob(repoName string, d digest.Digest) error

	LinkBlob(newRepo string, repo string, d digest.Digest) error
}
