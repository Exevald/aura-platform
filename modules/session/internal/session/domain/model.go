package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrInvalidSession  = errors.New("invalid session")
)

type SessionStatus string

const (
	SessionActive  SessionStatus = "active"
	SessionRevoked SessionStatus = "revoked"
	SessionExpired SessionStatus = "expired"
)

type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Status     SessionStatus
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	Persistent bool
}

func (s *Session) Available(now time.Time) bool {
	return s.Status == SessionActive && now.Before(s.ExpiresAt)
}

type SessionRepository interface {
	NextID() (uuid.UUID, error)
	Store(Session) error
	Find(uuid.UUID) (Session, error)
	FindActiveByUser(uuid.UUID) ([]Session, error)
}
