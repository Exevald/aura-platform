package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrCredentialNotFound = errors.New("credential not found")
	ErrCredentialExists   = errors.New("credential already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidCredential  = errors.New("invalid credential")
	ErrInvalidPassword    = errors.New("password must contain between 8 and 128 characters")
)

type AuthMethod string

const AuthMethodPassword AuthMethod = "password"

type Credential struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Login        string
	PasswordHash string
	Method       AuthMethod
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CredentialRepository interface {
	NextID() (uuid.UUID, error)
	Store(credential Credential) error
	Find(id uuid.UUID) (Credential, error)
	FindByLogin(login string) (Credential, error)
	FindByUserID(userID uuid.UUID) (Credential, error)
	Delete(id uuid.UUID) error
}
