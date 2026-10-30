package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	sessionevents "session/api/events"
)

type sessionClock struct {
	now time.Time
}

func (c *sessionClock) Now() time.Time {
	return c.now
}

type mockSessionRepository struct {
	sessions map[uuid.UUID]Session
}

func (r *mockSessionRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *mockSessionRepository) Store(session Session) error {
	r.sessions[session.ID] = session
	return nil
}

func (r *mockSessionRepository) Find(id uuid.UUID) (Session, error) {
	session, ok := r.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return session, nil
}

func (r *mockSessionRepository) FindActiveByUser(userID uuid.UUID) ([]Session, error) {
	var sessions []Session
	for _, session := range r.sessions {
		if session.UserID == userID && session.Status == SessionActive {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

type sessionDispatcherFake struct {
	messages []proto.Message
}

func (d *sessionDispatcherFake) Dispatch(message proto.Message) error {
	d.messages = append(d.messages, message)
	return nil
}

func TestSessionServiceUsesSelectedLifetime(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name       string
		persistent bool
		want       time.Duration
	}{
		{name: "browser session", want: 8 * time.Hour},
		{name: "remember me", persistent: true, want: 30 * 24 * time.Hour},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := &mockSessionRepository{sessions: make(map[uuid.UUID]Session)}
			dispatcher := &sessionDispatcherFake{}
			service := NewSessionService(repo, dispatcher, &sessionClock{now: now}, 8*time.Hour, 30*24*time.Hour)
			id, expiresAt, err := service.Create(uuid.New(), test.persistent)
			require.NoError(t, err)
			require.Equal(t, now.Add(test.want), expiresAt)
			require.Equal(t, test.persistent, repo.sessions[id].Persistent)
			require.IsType(t, &sessionevents.SessionCreated{}, dispatcher.messages[0])
		})
	}
}

func TestSessionServiceValidateAndRevokeAll(t *testing.T) {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	clock := &sessionClock{now: now}
	repo := &mockSessionRepository{sessions: make(map[uuid.UUID]Session)}
	dispatcher := &sessionDispatcherFake{}
	service := NewSessionService(repo, dispatcher, clock, time.Hour, 24*time.Hour)
	userID := uuid.New()
	id, _, err := service.Create(userID, false)
	require.NoError(t, err)

	_, err = service.Validate(id)
	require.NoError(t, err)
	clock.now = now.Add(2 * time.Hour)
	_, err = service.Validate(id)
	require.ErrorIs(t, err, ErrInvalidSession)

	clock.now = now.Add(time.Minute)
	err = service.RevokeAll(userID)
	require.NoError(t, err)
	require.Equal(t, SessionRevoked, repo.sessions[id].Status)
	require.IsType(t, &sessionevents.SessionRevoked{}, dispatcher.messages[1])
}
