package filesystem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/opencontainers/go-digest"
)

var (
	rootPath  = os.Getenv("STORAGE_PATH")
	uploadDir = filepath.Join(rootPath, "uploads")
	blobDir   = filepath.Join(rootPath, "blobs")
)

func initDirs() error {
	dirs := []string{rootPath, uploadDir, blobDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

type Storage struct {
}

func NewStorage() *Storage {
	err := initDirs()
	if err != nil {
		panic(err)
	}
	return &Storage{}
}

func (s *Storage) GetUploadBlobSize(id string) (int64, error) {
	tmpPath := filepath.Join(uploadDir, id)
	st, err := os.Stat(tmpPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, storage.ErrNotFound
		}
		return 0, storage.ErrStorageFail
	}
	return st.Size(), nil
}

func (s *Storage) UploadBlob(id string, r io.Reader) (int64, error) {
	tmpPath := filepath.Join(uploadDir, id)
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed to open upload file %s: %+v", tmpPath, err)
		return 0, fmt.Errorf("failed to open upload file: %w", storage.ErrStorageFail)
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	if err != nil {
		log.Printf("failed to save file content to %s: %+v", tmpPath, err)
		return 0, fmt.Errorf("failed to save content: %w", storage.ErrStorageFail)
	}

	stat, err := os.Stat(tmpPath)
	if err != nil {
		log.Printf("failed to get tmp file stat %s: %+v", tmpPath, err)
		return 0, fmt.Errorf("failed to get tmp file stat: %w", storage.ErrStorageFail)
	}

	return stat.Size(), nil
}

func (s *Storage) CommitBlob(repoName string, id string, d digest.Digest) error {
	tmpPath := filepath.Join(uploadDir, id)
	tmpFile, err := os.Open(tmpPath)
	if err != nil {
		log.Printf("failed to open upload file %s: %+v", tmpPath, err)
		return fmt.Errorf("failed to open upload file: %w", storage.ErrStorageFail)
	}
	defer tmpFile.Close()

	verifier := d.Verifier()
	_, err = io.Copy(verifier, tmpFile)
	if err != nil {
		log.Printf("failed to verify digest for file %s: %+v", tmpPath, err)
		return fmt.Errorf("failed to verify: %w", storage.ErrStorageFail)
	}
	if !verifier.Verified() {
		log.Printf("not verified digest: %s", d.String())
		return storage.ErrNotVerified
	}

	if err = os.MkdirAll(filepath.Join(blobDir, repoName), 0755); err != nil {
		return fmt.Errorf("failed to create blob dir: %w", storage.ErrStorageFail)
	}

	blobPath := filepath.Join(blobDir, repoName, d.String())
	if err := os.Rename(tmpPath, blobPath); err != nil {
		log.Printf("failed to store blob %s: %+v", blobPath, err)
		return fmt.Errorf("failed to store blob: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) SaveBlob(repoName string, d digest.Digest, data []byte) error {
	if err := os.MkdirAll(filepath.Join(blobDir, repoName), 0755); err != nil {
		return fmt.Errorf("failed to create blob dir: %w", storage.ErrStorageFail)
	}

	manifestPath := filepath.Join(blobDir, repoName, d.String())
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to store manifest: %w", storage.ErrStorageFail)
	}

	return nil
}

func (s *Storage) ReadBlob(repoName string, d digest.Digest) ([]byte, error) {
	blobPath := filepath.Join(blobDir, repoName, d.String())
	data, err := os.ReadFile(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, storage.ErrNotFound
		}

		return []byte{}, fmt.Errorf("failed to read manifest file: %w", storage.ErrStorageFail)
	}

	return data, nil
}

func (s *Storage) ReadBlobToWriter(repoName string, d digest.Digest, w io.Writer) (int64, error) {
	blobPath := filepath.Join(blobDir, repoName, d.String())
	st, err := os.Stat(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, storage.ErrNotFound
		}
		return 0, fmt.Errorf("failed to stat: %w", storage.ErrStorageFail)
	}

	f, err := os.Open(blobPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open upload file: %w", storage.ErrStorageFail)
	}
	defer f.Close()

	verifier := d.Verifier()
	writer := io.MultiWriter(w, verifier)

	_, err = io.Copy(writer, f)
	if err != nil {
		return 0, fmt.Errorf("failed to copy blob: %w", storage.ErrStorageFail)
	}

	if !verifier.Verified() {
		return 0, storage.ErrNotVerified
	}

	return st.Size(), nil
}

func (s *Storage) IsExistBlob(repoName string, d digest.Digest) (bool, error) {
	path := filepath.Join(blobDir, repoName, d.String())
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, storage.ErrStorageFail
	}
	return true, nil
}

func (s *Storage) DeleteBlob(repoName string, d digest.Digest) error {
	path := filepath.Join(blobDir, repoName, d.String())
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return fmt.Errorf("failed to remove blob: %w", storage.ErrStorageFail)
	}
	return nil
}

func (s *Storage) LinkBlob(newRepo string, repo string, d digest.Digest) error {
	oldPath := filepath.Join(blobDir, repo, d.String())
	newPath := filepath.Join(blobDir, newRepo, d.String())
	if _, err := os.Stat(oldPath); err != nil {
		if os.IsNotExist(err) {
			return storage.ErrNotFound
		}
		return fmt.Errorf("failed to stat blob: %w", storage.ErrStorageFail)
	}

	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return fmt.Errorf("failed to create repository: %w", storage.ErrStorageFail)
	}

	if err := os.Link(oldPath, newPath); err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("failed to create hard link: %w", storage.ErrStorageFail)
	}

	return nil
}
