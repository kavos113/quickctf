package storage

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrStorageFail  = errors.New("file storage has problem")
	ErrNotVerified  = errors.New("not verified")
	ErrInvalidRange = errors.New("invalid range")
)
