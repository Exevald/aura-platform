package domain

import (
	"errors"
	"uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidUser  = errors.New("invalid user")
	ErrEmailTaken   = errors.New("email is already in use")
)

type User struct {
	ID          uuid.UUID
	DisplayName string
	FirstName   string
	LastName    string
	Email       string
}

type UserRepository interface {
	NextID() (uuid.UUID, error)
	Store(User) error
	Find(id uuid.UUID) (User, error)
	FindByEmail(email string) (User, error)
	Delete(id uuid.UUID) error
}
