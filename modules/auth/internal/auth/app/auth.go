package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	"auth/internal/auth/domain"
	"auth/internal/auth/infra/mysql"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrEmailTaken      = errors.New("email is already in use")
)

type ProfileSpec struct {
	DisplayName string
	FirstName   string
	LastName    string
	Email       string
}

type DispatcherFactory interface {
	NewDispatcher(ctx context.Context) event.Dispatcher
}

type UnitOfWork interface {
	ExecuteWithRepositoryProvider(context.Context, func(*mysql.RepositoryProvider) error) error
}

type PasswordHasher = domain.PasswordHasher

type UserClient interface {
	CreateUser(context.Context, ProfileSpec) (uuid.UUID, error)
	DeleteUser(context.Context, uuid.UUID) error
}

type SessionClient interface {
	CreateSession(context.Context, uuid.UUID, bool) (uuid.UUID, time.Time, error)
	ValidateSession(context.Context, uuid.UUID) (uuid.UUID, time.Time, error)
	RevokeSession(context.Context, uuid.UUID) error
	RevokeUserSessions(context.Context, uuid.UUID) error
}

type AccessClient interface {
	Grant(context.Context, uuid.UUID) error
	Revoke(context.Context, uuid.UUID) error
}

type RegisterSpec struct {
	Login      string
	Password   string
	Profile    ProfileSpec
	RememberMe bool
}

type AuthResult struct {
	UserID        uuid.UUID
	SessionID     uuid.UUID
	ExpiresAt     time.Time
	Authenticated bool
	Persistent    bool
}

type AuthService interface {
	Register(context.Context, RegisterSpec) (AuthResult, error)
	Login(context.Context, string, string, bool) (AuthResult, error)
	Logout(context.Context, uuid.UUID) error
	ChangePassword(context.Context, uuid.UUID, string, string) error
}

func NewAuthService(
	unit UnitOfWork,
	dispatcherFactory DispatcherFactory,
	hasher PasswordHasher,
	users UserClient,
	sessions SessionClient,
	access AccessClient,
) AuthService {
	return &authService{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
		hasher:            hasher,
		users:             users,
		sessions:          sessions,
		access:            access,
		clock:             domain.SystemClock{},
	}
}

type authService struct {
	unit              UnitOfWork
	dispatcherFactory DispatcherFactory
	hasher            PasswordHasher
	users             UserClient
	sessions          SessionClient
	access            AccessClient
	clock             domain.Clock
}

func (s *authService) Register(ctx context.Context, spec RegisterSpec) (AuthResult, error) {
	if spec.Login == "" {
		return AuthResult{}, domain.ErrInvalidCredential
	}
	err := s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).EnsureLoginAvailable(spec.Login)
	})
	if err != nil {
		return AuthResult{}, err
	}
	userID, err := s.users.CreateUser(ctx, spec.Profile)
	if err != nil {
		return AuthResult{}, err
	}
	if err = s.access.Grant(ctx, userID); err != nil {
		_ = s.users.DeleteUser(ctx, userID)
		return AuthResult{}, err
	}

	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		_, createErr := s.domainService(ctx, provider).CreateCredential(userID, spec.Login, spec.Password)
		return createErr
	})
	if err != nil {
		_ = s.access.Revoke(ctx, userID)
		_ = s.users.DeleteUser(ctx, userID)
		return AuthResult{}, err
	}

	result := AuthResult{UserID: userID}
	result.SessionID, result.ExpiresAt, err = s.sessions.CreateSession(ctx, userID, spec.RememberMe)
	if err != nil {
		return result, nil
	}
	result.Authenticated = true
	result.Persistent = spec.RememberMe
	return result, nil
}

func (s *authService) Login(ctx context.Context, login, password string, rememberMe bool) (AuthResult, error) {
	var credential domain.Credential
	err := s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		var authErr error
		credential, authErr = s.domainService(ctx, provider).Authenticate(login, password)
		return authErr
	})
	if err != nil {
		return AuthResult{}, err
	}

	sessionID, expiresAt, err := s.sessions.CreateSession(ctx, credential.UserID, rememberMe)
	if err != nil {
		return AuthResult{}, err
	}
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).RecordLogin(credential)
	})
	if err != nil {
		_ = s.sessions.RevokeSession(ctx, sessionID)
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
	err := s.sessions.RevokeSession(ctx, sessionID)
	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}
	return err
}

func (s *authService) ChangePassword(ctx context.Context, sessionID uuid.UUID, current, next string) error {
	userID, _, err := s.sessions.ValidateSession(ctx, sessionID)
	if err != nil {
		return err
	}
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(provider *mysql.RepositoryProvider) error {
		return s.domainService(ctx, provider).ChangePassword(userID, current, next)
	})
	if err != nil {
		return err
	}
	return s.sessions.RevokeUserSessions(ctx, userID)
}

func (s *authService) domainService(ctx context.Context, provider *mysql.RepositoryProvider) domain.AuthService {
	return domain.NewAuthService(
		provider.Credential(ctx),
		s.hasher,
		s.dispatcherFactory.NewDispatcher(ctx),
		s.clock,
	)
}
