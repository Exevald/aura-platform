package app

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
)

type fakeSSOSessions struct{ userID uuid.UUID }

func (f fakeSSOSessions) CreateSession(context.Context, uuid.UUID, bool) (uuid.UUID, time.Time, error) {
	return uuid.New(), time.Now().Add(time.Hour), nil
}
func (f fakeSSOSessions) ValidateSession(context.Context, uuid.UUID) (uuid.UUID, time.Time, error) {
	return f.userID, time.Now().Add(time.Hour), nil
}
func (f fakeSSOSessions) RevokeSession(context.Context, uuid.UUID) error      { return nil }
func (f fakeSSOSessions) RevokeUserSessions(context.Context, uuid.UUID) error { return nil }

type fakeSSOTickets struct {
	hash    []byte
	userID  uuid.UUID
	target  string
	expires time.Time
}

func (f *fakeSSOTickets) Save(_ context.Context, hash []byte, userID uuid.UUID, target string, expires time.Time) error {
	f.hash, f.userID, f.target, f.expires = append([]byte(nil), hash...), userID, target, expires
	return nil
}
func (f *fakeSSOTickets) Take(_ context.Context, hash []byte, target string) (uuid.UUID, error) {
	if f.target != target || !time.Now().Before(f.expires) || string(f.hash) != string(hash) {
		return uuid.Nil(), ErrInvalidSSOTicket
	}
	f.hash = nil
	return f.userID, nil
}

func TestSSOTicketIsBoundToTargetAndSingleUse(t *testing.T) {
	userID := uuid.New()
	store := &fakeSSOTickets{}
	svc := NewSSOService(fakeSSOSessions{userID: userID}, store)
	code, err := svc.Issue(t.Context(), uuid.New(), "aura-platform.documents")
	require.NoError(t, err)
	_, err = svc.Consume(t.Context(), code, "aura-platform.calendar")
	require.ErrorIs(t, err, ErrInvalidSSOTicket)
	result, err := svc.Consume(t.Context(), code, "aura-platform.documents")
	require.NoError(t, err)
	require.Equal(t, userID, result.UserID)
	_, err = svc.Consume(t.Context(), code, "aura-platform.documents")
	require.ErrorIs(t, err, ErrInvalidSSOTicket)
}

func TestSSOTicketRejectsInvalidTargetsAndExpiredCodes(t *testing.T) {
	store := &fakeSSOTickets{}
	svc := NewSSOService(fakeSSOSessions{userID: uuid.New()}, store)
	_, err := svc.Issue(t.Context(), uuid.New(), "evil.example")
	require.ErrorIs(t, err, ErrInvalidSSOTarget)
	_, err = svc.Consume(t.Context(), "bad", "aura-platform.home")
	require.ErrorIs(t, err, ErrInvalidSSOTicket)
	code, err := svc.Issue(t.Context(), uuid.New(), "aura-platform.home")
	require.NoError(t, err)
	store.expires = time.Now().Add(-time.Second)
	_, err = svc.Consume(t.Context(), code, "aura-platform.home")
	require.True(t, errors.Is(err, ErrInvalidSSOTicket))
}
