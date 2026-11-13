package model

import (
	"errors"
	"time"
	"uuid"
)

const (
	SharePending      ShareStatus = "pending"
	ShareActive       ShareStatus = "active"
	ShareCompensating ShareStatus = "compensating"
	ShareFailed       ShareStatus = "failed"
	ShareRevoked      ShareStatus = "revoked"
)

var (
	ErrShareNotFound = errors.New("group share not found")
	ErrShareConflict = errors.New("group share already exists")
	ErrInvalidShare  = errors.New("invalid group share")
)

type ShareStatus string
type ModuleShare struct {
	ID, ModuleID, GroupID, OwnerUserID uuid.UUID
	Status                             ShareStatus
	Attempt                            int
	FailureReason                      string
	CreatedAt, UpdatedAt               time.Time
}

type ModuleShareRepository interface {
	NextID() (uuid.UUID, error)
	Store(ModuleShare) error
	Find(uuid.UUID) (ModuleShare, error)
	FindByOwnerGroup(uuid.UUID, uuid.UUID) (ModuleShare, error)
	ListByOwner(uuid.UUID) ([]ModuleShare, error)
	ListPendingBefore(time.Time) ([]ModuleShare, error)
	ListCompensatingBefore(time.Time) ([]ModuleShare, error)
}
