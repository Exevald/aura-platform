package domain

import (
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	sessionevents "session/api/events"
)

type SessionService interface {
	Create(userID uuid.UUID, persistent bool) (uuid.UUID, time.Time, error)
	Validate(id uuid.UUID) (Session, error)
	Revoke(id uuid.UUID) error
	RevokeAll(userID uuid.UUID) error
}

func NewSessionService(
	repo SessionRepository,
	dispatcher event.Dispatcher,
	clock Clock,
	defaultLifetime time.Duration,
	persistentLifetime time.Duration,
) SessionService {
	return &sessionService{
		repo:               repo,
		dispatcher:         dispatcher,
		clock:              clock,
		defaultLifetime:    defaultLifetime,
		persistentLifetime: persistentLifetime,
	}
}

type sessionService struct {
	repo               SessionRepository
	dispatcher         event.Dispatcher
	clock              Clock
	defaultLifetime    time.Duration
	persistentLifetime time.Duration
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

func (s *sessionService) Create(userID uuid.UUID, persistent bool) (uuid.UUID, time.Time, error) {
	lifetime := s.defaultLifetime
	if persistent {
		lifetime = s.persistentLifetime
	}
	if userID == uuid.Nil() || lifetime <= 0 {
		return uuid.Nil(), time.Time{}, ErrInvalidSession
	}

	id, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}

	now := s.clock.Now().UTC()
	session := Session{
		ID:         id,
		UserID:     userID,
		Status:     SessionActive,
		CreatedAt:  now,
		ExpiresAt:  now.Add(lifetime),
		Persistent: persistent,
	}
	err = s.repo.Store(session)
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}

	err = s.dispatcher.Dispatch(&sessionevents.SessionCreated{
		Id:             session.ID.String(),
		UserId:         session.UserID.String(),
		OccurredAtUnix: now.Unix(),
		ExpiresAtUnix:  session.ExpiresAt.Unix(),
		Persistent:     session.Persistent,
	})
	if err != nil {
		return uuid.Nil(), time.Time{}, err
	}

	return id, session.ExpiresAt, nil
}

func (s *sessionService) Validate(id uuid.UUID) (Session, error) {
	session, err := s.repo.Find(id)
	if err != nil {
		return Session{}, err
	}
	if !session.Available(s.clock.Now().UTC()) {
		return Session{}, ErrInvalidSession
	}

	return session, nil
}

func (s *sessionService) Revoke(id uuid.UUID) error {
	session, err := s.repo.Find(id)
	if err != nil {
		return err
	}

	return s.revokeSession(session)
}

func (s *sessionService) RevokeAll(userID uuid.UUID) error {
	sessions, err := s.repo.FindActiveByUser(userID)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		err = s.revokeSession(session)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *sessionService) revokeSession(session Session) error {
	if session.Status == SessionRevoked {
		return nil
	}

	now := s.clock.Now().UTC()
	session.Status, session.RevokedAt = SessionRevoked, &now
	err := s.repo.Store(session)
	if err != nil {
		return err
	}

	return s.dispatcher.Dispatch(&sessionevents.SessionRevoked{
		Id:             session.ID.String(),
		UserId:         session.UserID.String(),
		OccurredAtUnix: now.Unix(),
	})
}
