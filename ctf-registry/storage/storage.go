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
	CommitBlob(id string, d digest.Digest) error
	SaveBlob(d digest.Digest, data []byte) error
	ReadBlob(d digest.Digest) ([]byte, error)
	// ReadBlobToWriter write blob to w, and verify digest
	ReadBlobToWriter(d digest.Digest, w io.Writer) (int64, error)
	IsExistBlob(d digest.Digest) (bool, error)
	DeleteBlob(d digest.Digest) error
}
