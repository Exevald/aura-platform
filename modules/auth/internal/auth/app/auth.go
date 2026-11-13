package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	"auth/internal/auth/domain"
)

var ErrSessionNotFound = errors.New("session not found")

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) event.Dispatcher
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type RepositoryProvider interface {
	Credential(context.Context) domain.CredentialRepository
}

type PasswordHasher = domain.PasswordHasher

type SessionClient interface {
	CreateSession(context.Context, uuid.UUID, bool) (uuid.UUID, time.Time, error)
	ValidateSession(context.Context, uuid.UUID) (uuid.UUID, time.Time, error)
	RevokeSession(context.Context, uuid.UUID) error
	RevokeUserSessions(context.Context, uuid.UUID) error
}

type AuthResult struct {
	UserID        uuid.UUID
	SessionID     uuid.UUID
	ExpiresAt     time.Time
	Authenticated bool
	Persistent    bool
}

type AuthService interface {
	Login(context.Context, string, string, bool) (AuthResult, error)
	Logout(context.Context, uuid.UUID) error
	ChangePassword(context.Context, uuid.UUID, string, string) error
}

type authService struct {
	unit              UnitOfWork
	dispatcherFactory DispatcherFactory
	hasher            PasswordHasher
	sessionAdapter    SessionClient
	clock             domain.Clock
}

func NewAuthService(
	unit UnitOfWork,
	dispatcherFactory DispatcherFactory,
	hasher PasswordHasher,
	sessions SessionClient,
) AuthService {
	return &authService{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
		hasher:            hasher,
		sessionAdapter:    sessions,
		clock:             domain.SystemClock{},
	}
}

func (s *authService) Login(ctx context.Context, login, password string, rememberMe bool) (AuthResult, error) {
	var credential domain.Credential
	err := s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		var authErr error
		credential, authErr = s.domainService(ctx, provider).Authenticate(login, password)
		return authErr
	})
	if err != nil {
		return AuthResult{}, err
	}
	sessionID, expiresAt, err := s.sessionAdapter.CreateSession(ctx, credential.UserID, rememberMe)
	if err != nil {
		return AuthResult{}, err
	}
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider).RecordLogin(credential)
	})
	if err != nil {
		_ = s.sessionAdapter.RevokeSession(ctx, sessionID)
		return AuthResult{}, err
	}
	return AuthResult{
		UserID:        credential.UserID,
		SessionID:     sessionID,
		ExpiresAt:     expiresAt,
		Authenticated: true,
		Persistent:    rememberMe,
	}, nil
}

func (s *authService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if sessionID == uuid.Nil() {
		return nil
	}
	err := s.sessionAdapter.RevokeSession(ctx, sessionID)
	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}
	return err
}

func (s *authService) ChangePassword(ctx context.Context, sessionID uuid.UUID, current, next string) error {
	userID, _, err := s.sessionAdapter.ValidateSession(ctx, sessionID)
	if err != nil {
		return err
	}
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return s.domainService(ctx, provider).ChangePassword(userID, current, next)
	})
	if err != nil {
		return err
	}
	return s.sessionAdapter.RevokeUserSessions(ctx, userID)
}

func (s *authService) domainService(
	ctx context.Context,
	provider RepositoryProvider,
) domain.AuthService {
	return domain.NewAuthService(
		provider.Credential(ctx),
		s.hasher,
		s.dispatcherFactory.NewDispatcher(ctx),
		s.clock,
	)
}
