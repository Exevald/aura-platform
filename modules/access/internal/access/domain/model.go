package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrModuleAccessNotFound = errors.New("module access not found")
	ErrInvalidModuleAccess  = errors.New("invalid module access")
)

type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

type ModuleAccess struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ModuleID  uuid.UUID
	Effect    Effect
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ModuleAccessRepository interface {
	NextID() (uuid.UUID, error)
	Store(ModuleAccess) error
	Find(uuid.UUID) (ModuleAccess, error)
	FindByUserAndModule(userID, moduleID uuid.UUID) ([]ModuleAccess, error)
	Delete(uuid.UUID) error
}
