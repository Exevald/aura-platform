package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalidTarget = errors.New("target URL must be an absolute http or https URL")
	ErrNotFound      = errors.New("short link not found")
	ErrCodeConflict  = errors.New("short code conflict")
)

type Link struct {
	ID        uuid.UUID
	OwnerID   uuid.UUID
	Code      string
	TargetURL string
	CreatedAt time.Time
}

type LinkRepository interface {
	NextID() (uuid.UUID, error)
	Store(Link) error
	FindByID(uuid.UUID) (Link, error)
	ListByOwner(uuid.UUID) ([]Link, error)
	FindByCode(string) (Link, error)
	Delete(uuid.UUID, uuid.UUID) error
}
