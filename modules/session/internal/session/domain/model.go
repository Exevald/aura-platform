package domain

import (
	"errors"
	"time"
	"uuid"
)

const (
	SessionActive  SessionStatus = "active"
	SessionRevoked SessionStatus = "revoked"
	SessionExpired SessionStatus = "expired"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidSession  = errors.New("invalid session")
)

type SessionStatus string

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Status     SessionStatus
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	Persistent bool
}

type SessionRepository interface {
	NextID() (uuid.UUID, error)
	Store(Session) error
	Find(uuid.UUID) (Session, error)
	FindActiveByUser(uuid.UUID) ([]Session, error)
}

func (s *Session) Available(now time.Time) bool {
	return s.Status == SessionActive && now.Before(s.ExpiresAt)
}
