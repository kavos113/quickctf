package boltstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/kavos113/quickctf/ctf-registry/manifest"
	"github.com/kavos113/quickctf/ctf-registry/storage"
	"github.com/opencontainers/go-digest"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db *bolt.DB
}

var (
	bucketNameTag        = []byte("tag")
	bucketNameReference  = []byte("reference")
	bucketNameSession    = []byte("session")
	bucketNameRepository = []byte("repository")
)

func NewStore() *Storage {
	_db, err := bolt.Open(filepath.Join(os.Getenv("STORAGE_PATH"), "minicr.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = _db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketNameTag)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(bucketNameReference)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(bucketNameSession)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(bucketNameRepository)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{db: _db}
}

func (s Storage) SaveTag(repoName string, d digest.Digest, tag string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameTag)
		err := b.Put(tagKey(repoName, tag), []byte(d.String()))
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to save tag: %w", storage.ErrStorageFail)
	}
	return nil
}

func (s Storage) ReadTag(repoName string, tag string) (string, error) {
	var value string
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameTag)
		v := b.Get(tagKey(repoName, tag))
		if v == nil {
			return storage.ErrNotFound
		}
		value = string(v)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

func (s Storage) DeleteTag(repoName string, tag string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameTag)
		v := b.Get(tagKey(repoName, tag))
		if v == nil {
			return storage.ErrNotFound
		}
		err := b.Delete(tagKey(repoName, tag))
		return err
	})
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete tag: %w", storage.ErrStorageFail)
	}
	return nil
}

func (s Storage) GetTagList(repoName string, limit int, last string) ([]string, error) {
	tags := make([]string, 0)
	_ = s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(bucketNameTag).Cursor()

		prefix := []byte(fmt.Sprintf("%s:", repoName))
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			tags = append(tags, string(k))
		}
		return nil
	})

	if last != "" {
		i := slices.Index(tags, last)
		if i >= 0 && i < len(tags)-1 {
			tags = tags[i+1:]
		}
	}
	if limit > 0 {
		tags = tags[:limit]
	}

	return tags, nil
}

func tagKey(repoName string, tag string) []byte {
	return []byte(fmt.Sprintf("%s:%s", repoName, tag))
}

func (s Storage) AddReference(repoName string, d digest.Digest, desc manifest.Descriptor) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameReference)
		key := []byte(fmt.Sprintf("%s:%s", repoName, d.String()))
		curr := b.Get(key)

		var list []manifest.Descriptor

		if curr != nil {
			if err := json.Unmarshal(curr, &list); err != nil {
				return fmt.Errorf("broken descriptor: %w", storage.ErrStorageFail)
			}
		}

		list = append(list, desc)
		data, err := json.Marshal(list)
		if err != nil {
			return fmt.Errorf("broken data: %w", storage.ErrStorageFail)
		}

		err = b.Put(key, data)
		if err != nil {
			return fmt.Errorf("failed to put storage: %w", storage.ErrStorageFail)
		}
		return nil
	})
}

func (s Storage) GetReferences(repoName string, d digest.Digest, artifactType string) ([]manifest.Descriptor, error) {
	var list []manifest.Descriptor

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameReference)
		key := []byte(fmt.Sprintf("%s:%s", repoName, d.String()))
		data := b.Get(key)
		if data == nil {
			return storage.ErrNotFound
		}

		if err := json.Unmarshal(b.Get(key), &list); err != nil {
			return fmt.Errorf("broken descriptor: %w", storage.ErrStorageFail)
		}
		return nil
	})

	if artifactType == "" {
		return list, err
	}

	filtered := make([]manifest.Descriptor, 0)
	for _, d := range list {
		fmt.Printf("desc type: %s, digest: %s\n", *d.ArtifactType, d.Digest.String())
		if *d.ArtifactType == artifactType {
			filtered = append(filtered, d)
		}
	}
	return filtered, err
}

func (s Storage) AddBlob(repoName string, d digest.Digest) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameRepository)
		key := []byte(fmt.Sprintf("%s@%s", repoName, d.String()))
		return b.Put(key, []byte{1})
	})
	if err != nil {
		return fmt.Errorf("failed to add blob: %w", storage.ErrStorageFail)
	}
	return nil
}

func (s Storage) DeleteBlob(repoName string, d digest.Digest) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameRepository)
		key := []byte(fmt.Sprintf("%s@%s", repoName, d.String()))
		v := b.Get(key)
		if v == nil {
			return storage.ErrNotFound
		}

		err := b.Delete(key)
		if err != nil {
			return fmt.Errorf("failed to delete blob: %w", storage.ErrStorageFail)
		}
		return nil
	})
}

func (s Storage) IsExistBlob(repoName string, d digest.Digest) (bool, error) {
	var exists bool
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameRepository)
		key := []byte(fmt.Sprintf("%s@%s", repoName, d.String()))
		v := b.Get(key)
		exists = v != nil
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to check blob: %w", storage.ErrStorageFail)
	}
	return exists, nil
}

func (s Storage) LinkBlob(newRepo string, d digest.Digest) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketNameRepository)
		key := []byte(fmt.Sprintf("%s@%s", newRepo, d.String()))
		return b.Put(key, []byte{1})
	})
}
