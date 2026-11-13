package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
	"uuid"
)

var (
	ErrGroupNotFound = errors.New("group not found")
	ErrInvalidGroup  = errors.New("invalid group")
)

type Group struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	WorkspaceID uuid.UUID
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GroupRepository interface {
	NextID() (uuid.UUID, error)
	Store(Group) error
	Find(uuid.UUID) (Group, error)
	ListByOwner(uuid.UUID, uuid.UUID) ([]Group, error)
	Delete(uuid.UUID) error
}

func validName(name string) bool {
	name = strings.TrimSpace(name)
	return name != "" && utf8.ValidString(name) && utf8.RuneCountInString(name) <= 120
}
