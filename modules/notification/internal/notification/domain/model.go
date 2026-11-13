package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalid  = errors.New("invalid notification")
	ErrNotFound = errors.New("notification not found")
)

type Notification struct {
	ID, UserID, WorkspaceID           uuid.UUID
	IdempotencyKey, Kind, Title, Body string
	RecipientEmail                    string
	Read, EmailRequested              bool
	CreatedAt                         time.Time
}

type Preferences struct {
	UserID       uuid.UUID
	InApp, Email bool
}

type Repository interface {
	NextID() (uuid.UUID, error)
	Store(Notification) (bool, error)
	List(uuid.UUID) ([]Notification, error)
	MarkRead(uuid.UUID, uuid.UUID) error
	GetPreferences(uuid.UUID) (Preferences, error)
	StorePreferences(Preferences) error
}
