package app

import (
	"context"
	"errors"
	"time"
	"uuid"
)

const (
	RegistrationPreparing  RegistrationState = "preparing"
	RegistrationCommitting RegistrationState = "committing"
	RegistrationAborting   RegistrationState = "aborting"
	RegistrationCommitted  RegistrationState = "committed"
	RegistrationAborted    RegistrationState = "aborted"
)

const (
	registrationLeaseDuration = 30 * time.Second
	expiredRegistrationError  = "coordinator lease expired"
	authBranch                = registrationBranch("auth")
	userBranch                = registrationBranch("user")
	accessBranch              = registrationBranch("access")
)

var (
	ErrInvalidRegistration = errors.New("invalid registration")
	ErrLoginTaken          = errors.New("login is already in use")
	ErrEmailTaken          = errors.New("email is already in use")
)

var errInvalidRegistrationTransition = errors.New("invalid registration transaction transition")

type RegistrationState string

type RegistrationTransaction struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CredentialID uuid.UUID
	Login        string
	State        RegistrationState
	AuthDone     bool
	UserDone     bool
	AccessDone   bool
	LeaseUntil   time.Time
	LastError    string
}

type RegistrationStorage interface {
	Save(RegistrationTransaction) error
	FindForUpdate(uuid.UUID) (RegistrationTransaction, error)
	FindPending(time.Time, int) ([]RegistrationTransaction, error)
}

type StorageProvider interface {
	Registration(context.Context) RegistrationStorage
}

type registrationBranch string

func (t *RegistrationTransaction) decideCommit(credentialID uuid.UUID) error {
	if t.State != RegistrationPreparing || credentialID == uuid.Nil() {
		return errInvalidRegistrationTransition
	}
	t.State = RegistrationCommitting
	t.CredentialID = credentialID
	t.LeaseUntil = time.Time{}
	t.LastError = ""
	return nil
}

func (t *RegistrationTransaction) decideAbort(cause error) error {
	if t.State == RegistrationAborting {
		return nil
	}
	if t.State != RegistrationPreparing || cause == nil {
		return errInvalidRegistrationTransition
	}
	t.State = RegistrationAborting
	t.CredentialID = uuid.Nil()
	t.LeaseUntil = time.Time{}
	t.LastError = cause.Error()
	return nil
}

func (t *RegistrationTransaction) expire(now time.Time) (bool, error) {
	if t.State != RegistrationPreparing || t.LeaseUntil.IsZero() || !t.LeaseUntil.Before(now) {
		return false, nil
	}
	return true, t.decideAbort(errors.New(expiredRegistrationError))
}

func (t *RegistrationTransaction) resolveBranch(branch registrationBranch) error {
	if t.State != RegistrationCommitting && t.State != RegistrationAborting {
		return errInvalidRegistrationTransition
	}
	switch branch {
	case authBranch:
		t.AuthDone = true
	case userBranch:
		t.UserDone = true
	case accessBranch:
		t.AccessDone = true
	default:
		return errInvalidRegistrationTransition
	}
	return nil
}

func (t *RegistrationTransaction) complete() (bool, error) {
	if !t.resolved() {
		return false, nil
	}
	switch t.State {
	case RegistrationCommitting:
		if t.CredentialID == uuid.Nil() {
			return false, errInvalidRegistrationTransition
		}
		t.State = RegistrationCommitted
		t.LastError = ""
		return true, nil
	case RegistrationAborting:
		t.State = RegistrationAborted
		t.LastError = ""
		return false, nil
	case RegistrationCommitted, RegistrationAborted:
		return false, nil
	default:
		return false, errInvalidRegistrationTransition
	}
}

func (t *RegistrationTransaction) decision() (bool, error) {
	switch t.State {
	case RegistrationCommitting:
		return true, nil
	case RegistrationAborting:
		return false, nil
	default:
		return false, errInvalidRegistrationTransition
	}
}

func (t *RegistrationTransaction) resolved() bool {
	return t.AuthDone && t.UserDone && t.AccessDone
}

func (t *RegistrationTransaction) terminal() bool {
	return t.State == RegistrationCommitted || t.State == RegistrationAborted
}

func newRegistrationTransaction(login string, now time.Time) RegistrationTransaction {
	return RegistrationTransaction{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		Login:      login,
		State:      RegistrationPreparing,
		LeaseUntil: now.Add(registrationLeaseDuration),
	}
}
