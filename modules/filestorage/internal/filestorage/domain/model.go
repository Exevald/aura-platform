package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalid   = errors.New("invalid file")
	ErrNotFound  = errors.New("file not found")
	ErrForbidden = errors.New("file access denied")
)

type File struct {
	ID, WorkspaceID, OwnerUserID                     uuid.UUID
	ObjectKey, FileName, ContentType, ChecksumSHA256 string
	SizeBytes                                        int64
	CreatedAt                                        time.Time
}

type Repository interface {
	Store(File) error
	Find(uuid.UUID) (File, error)
	List(uuid.UUID) ([]File, error)
	Delete(uuid.UUID) error
}
