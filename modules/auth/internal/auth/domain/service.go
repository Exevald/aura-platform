package domain

import (
	"errors"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	authevents "auth/api/events"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 128
)

type PasswordHasher interface {
	Hash(string) (string, error)
	Compare(hash, password string) error
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type AuthService interface {
	CreateCredential(userID uuid.UUID, login, password string) (Credential, error)
	EnsureLoginAvailable(login string) error
	Authenticate(login, password string) (Credential, error)
	RecordLogin(credential Credential) error
	ChangePassword(userID uuid.UUID, current, next string) error
}

type authService struct {
	repo       CredentialRepository
	hasher     PasswordHasher
	dispatcher event.Dispatcher
	clock      Clock
}

func NewAuthService(
	repo CredentialRepository,
	hasher PasswordHasher,
	dispatcher event.Dispatcher,
	clock Clock,
) AuthService {
	return &authService{
		repo:       repo,
		hasher:     hasher,
		dispatcher: dispatcher,
		clock:      clock,
	}
}

func (s *authService) CreateCredential(userID uuid.UUID, login, password string) (Credential, error) {
	if userID == uuid.Nil() || login == "" {
		return Credential{}, ErrInvalidCredential
	}
	if err := validatePassword(password); err != nil {
		return Credential{}, err
	}
	if err := s.EnsureLoginAvailable(login); err != nil {
		return Credential{}, err
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return Credential{}, err
	}
	id, err := s.repo.NextID()
	if err != nil {
		return Credential{}, err
	}

	now := s.clock.Now().UTC()
	credential := Credential{
		ID:           id,
		UserID:       userID,
		Login:        login,
		PasswordHash: hash,
		Method:       AuthMethodPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err = s.repo.Store(credential); err != nil {
		return Credential{}, err
	}
	if err = s.dispatcher.Dispatch(&authevents.UserRegistered{
		CredentialId:   credential.ID.String(),
		UserId:         credential.UserID.String(),
		Login:          credential.Login,
		OccurredAtUnix: now.Unix(),
	}); err != nil {
		return Credential{}, err
	}

	return credential, nil
}

func (s *authService) EnsureLoginAvailable(login string) error {
	if login == "" {
		return ErrInvalidCredential
	}
	_, err := s.repo.FindByLogin(login)
	if err == nil {
		return ErrCredentialExists
	}
	if errors.Is(err, ErrCredentialNotFound) {
		return nil
	}

	return err
}

func (s *authService) Authenticate(login, password string) (Credential, error) {
	if login == "" || password == "" {
		return Credential{}, ErrInvalidCredentials
	}
	credential, err := s.repo.FindByLogin(login)
	if errors.Is(err, ErrCredentialNotFound) {
		return Credential{}, ErrInvalidCredentials
	}
	if err != nil {
		return Credential{}, err
	}
	if s.hasher.Compare(credential.PasswordHash, password) != nil {
		return Credential{}, ErrInvalidCredentials
	}
	return credential, nil
}

func (s *authService) RecordLogin(credential Credential) error {
	return s.dispatcher.Dispatch(&authevents.UserLoggedIn{
		CredentialId:   credential.ID.String(),
		UserId:         credential.UserID.String(),
		Login:          credential.Login,
		OccurredAtUnix: s.clock.Now().UTC().Unix(),
	})
}

func (s *authService) ChangePassword(userID uuid.UUID, current, next string) error {
	if err := validatePassword(next); err != nil {
		return err
	}
	credential, err := s.verifyPassword(userID, current)
	if err != nil {
		return err
	}
	hash, err := s.hasher.Hash(next)
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC()
	credential.PasswordHash = hash
	credential.UpdatedAt = now
	if err = s.repo.Store(credential); err != nil {
		return err
	}
	return s.dispatcher.Dispatch(&authevents.PasswordChanged{
		CredentialId:   credential.ID.String(),
		UserId:         credential.UserID.String(),
		Login:          credential.Login,
		OccurredAtUnix: now.Unix(),
	})
}

func (s *authService) verifyPassword(userID uuid.UUID, password string) (Credential, error) {
	if userID == uuid.Nil() || password == "" {
		return Credential{}, ErrInvalidCredentials
	}
	credential, err := s.repo.FindByUserID(userID)
	if errors.Is(err, ErrCredentialNotFound) {
		return Credential{}, ErrInvalidCredentials
	}
	if err != nil {
		return Credential{}, err
	}
	if s.hasher.Compare(credential.PasswordHash, password) != nil {
		return Credential{}, ErrInvalidCredentials
	}
	return credential, nil
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength || length > maxPasswordLength {
		return ErrInvalidPassword
	}
	return nil
}
