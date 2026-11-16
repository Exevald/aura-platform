package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"
	"uuid"
)

var ErrInvalidSSOTicket = errors.New("invalid or expired SSO ticket")
var ErrInvalidSSOTarget = errors.New("invalid SSO target")

var productHosts = map[string]bool{
	"aura-platform.home":          true,
	"aura-platform.documents":     true,
	"aura-platform.calendar":      true,
	"aura-platform.planner":       true,
	"aura-platform.shortlinks":    true,
	"aura-platform.notifications": true,
}

type SSOTicketStore interface {
	Save(context.Context, []byte, uuid.UUID, string, time.Time) error
	Take(context.Context, []byte, string) (uuid.UUID, error)
}

type SSOService struct {
	sessions SessionClient
	tickets  SSOTicketStore
}

func NewSSOService(sessions SessionClient, tickets SSOTicketStore) *SSOService {
	return &SSOService{sessions: sessions, tickets: tickets}
}

func (s *SSOService) Issue(ctx context.Context, sessionID uuid.UUID, target string) (string, error) {
	if !productHosts[target] {
		return "", ErrInvalidSSOTarget
	}
	if sessionID == uuid.Nil() {
		return "", ErrSessionNotFound
	}
	userID, _, err := s.sessions.ValidateSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", err
	}
	hash := sha256.Sum256(raw)
	if err = s.tickets.Save(ctx, hash[:], userID, target, time.Now().UTC().Add(60*time.Second)); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (s *SSOService) Consume(ctx context.Context, code, host string) (AuthResult, error) {
	if !productHosts[host] {
		return AuthResult{}, ErrInvalidSSOTarget
	}
	raw, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil || len(raw) != 32 {
		return AuthResult{}, ErrInvalidSSOTicket
	}
	hash := sha256.Sum256(raw)
	userID, err := s.tickets.Take(ctx, hash[:], host)
	if err != nil {
		return AuthResult{}, err
	}
	id, expiresAt, err := s.sessions.CreateSession(ctx, userID, false)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{UserID: userID, SessionID: id, ExpiresAt: expiresAt, Authenticated: true}, nil
}
