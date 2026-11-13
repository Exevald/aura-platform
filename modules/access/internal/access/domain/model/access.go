package model

import (
	"errors"
	"time"
	"uuid"
)

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

var (
	ErrModuleAccessNotFound = errors.New("module access not found")
	ErrInvalidModuleAccess  = errors.New("invalid module access")
)

type Effect string

type ModuleAccess struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ModuleID  uuid.UUID
	Effect    Effect
	SourceID  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ModuleAccessRepository interface {
	NextID() (uuid.UUID, error)
	Store(ModuleAccess) error
	Find(uuid.UUID) (ModuleAccess, error)
	FindByUserAndModule(userID, moduleID uuid.UUID) ([]ModuleAccess, error)
	DeleteBySource(string) error
	DeleteBySourceAndUser(string, uuid.UUID) error
	Delete(uuid.UUID) error
}
