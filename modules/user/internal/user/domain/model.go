package domain

import (
	"errors"
	"uuid"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Echo struct {
	ID   uuid.UUID
	Body string
}

type User struct {
	ID        uuid.UUID
	Login     string
	FirstName string
	LastName  string
	Email     string
}

type UserRepository interface {
	NextID() (uuid.UUID, error)
	Store(User) error
	Find(id uuid.UUID) (User, error)
	Delete(id uuid.UUID) error
}

type EchoRepository interface {
	NextID() (uuid.UUID, error)
	Store(Echo) error
}
