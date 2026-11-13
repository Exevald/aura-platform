package model

import (
	"errors"
	"time"
	"uuid"
)

const (
	ModuleShareGranted   ModuleShareStatus = "granted"
	ModuleShareCancelled ModuleShareStatus = "cancelled"
)

var (
	ErrInvalidGroupModuleAccess     = errors.New("invalid group module access")
	ErrModuleShareOperationNotFound = errors.New("module share operation not found")
	ErrEmptyGroup                   = errors.New("group has no members")
	ErrUnknownModule                = errors.New("unknown module")
)

type ModuleShareStatus string

type ModuleShare struct {
	ShareID     uuid.UUID
	ModuleID    uuid.UUID
	GroupID     uuid.UUID
	OwnerUserID uuid.UUID
	Status      ModuleShareStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ModuleShareRepository interface {
	ListByGroup(uuid.UUID) ([]ModuleShare, error)
	ListByOwner(uuid.UUID) ([]ModuleShare, error)
	Find(uuid.UUID) (ModuleShare, error)
	Store(ModuleShare) error
}
