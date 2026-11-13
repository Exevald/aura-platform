package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalid  = errors.New("invalid mail")
	ErrNotFound = errors.New("mail not found")
)

type Job struct {
	ID                                                          uuid.UUID
	IdempotencyKey, Recipient, Subject, Body, Status, LastError string
	Attempts                                                    int
	CreatedAt, UpdatedAt                                        time.Time
}

type Repository interface {
	NextID() (uuid.UUID, error)
	Create(Job) (Job, bool, error)
	Find(uuid.UUID) (Job, error)
	FindByKey(string) (Job, error)
	StoreResult(Job) error
}
