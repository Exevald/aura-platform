package app

import (
	"context"
	"errors"
	"sync"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	registererevents "registerer/api/events"
)

type ProfileSpec struct {
	DisplayName string
	FirstName   string
	LastName    string
	Email       string
}

type RegisterSpec struct {
	Login      string
	Password   string
	Profile    ProfileSpec
	RememberMe bool
}

type RegisterResult struct {
	UserID        uuid.UUID
	SessionID     uuid.UUID
	ExpiresAt     time.Time
	Authenticated bool
	Persistent    bool
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) event.Dispatcher
}

type UnitOfWork interface {
	ExecuteWithStorageProvider(context.Context, func(StorageProvider) error) error
}

type AuthAdapter interface {
	PrepareRegistration(context.Context, uuid.UUID, uuid.UUID, string, string) (uuid.UUID, error)
	CommitRegistration(context.Context, uuid.UUID) error
	RollbackRegistration(context.Context, uuid.UUID) error
}

type UserAdapter interface {
	PrepareRegistration(context.Context, uuid.UUID, uuid.UUID, ProfileSpec) error
	CommitRegistration(context.Context, uuid.UUID) error
	RollbackRegistration(context.Context, uuid.UUID) error
}

type SessionAdapter interface {
	CreateSession(context.Context, uuid.UUID, bool) (uuid.UUID, time.Time, error)
}

type WorkspaceAdapter interface {
	EnsurePersonal(context.Context, uuid.UUID, uuid.UUID, string) error
}

type AccessAdapter interface {
	PrepareRegistration(context.Context, uuid.UUID, uuid.UUID) error
	CommitRegistration(context.Context, uuid.UUID) error
	RollbackRegistration(context.Context, uuid.UUID) error
}

type Service struct {
	unit              UnitOfWork
	dispatcherFactory DispatcherFactory
	authAdapter       AuthAdapter
	userAdapter       UserAdapter
	sessionAdapter    SessionAdapter
	workspaceAdapter  WorkspaceAdapter
	accessAdapter     AccessAdapter
}

type branchResolution struct {
	branch registrationBranch
	done   bool
	action func() error
	err    error
}

func NewService(
	unit UnitOfWork,
	dispatcherFactory DispatcherFactory,
	authAdapter AuthAdapter,
	userAdapter UserAdapter,
	sessionAdapter SessionAdapter,
	extras ...any,
) *Service {
	service := &Service{
		unit:              unit,
		dispatcherFactory: dispatcherFactory,
		authAdapter:       authAdapter,
		userAdapter:       userAdapter,
		sessionAdapter:    sessionAdapter,
	}
	for _, extra := range extras {
		switch v := extra.(type) {
		case WorkspaceAdapter:
			service.workspaceAdapter = v
		case AccessAdapter:
			service.accessAdapter = v
		}
	}
	return service
}

func (s *Service) Register(ctx context.Context, spec RegisterSpec) (RegisterResult, error) {
	if spec.Login == "" || spec.Password == "" {
		return RegisterResult{}, ErrInvalidRegistration
	}

	transaction, err := s.startRegistration(ctx, spec)
	if err != nil {
		return RegisterResult{}, err
	}

	credentialID, err := s.prepareRegistration(ctx, transaction, spec)
	if err != nil {
		s.abortRegistration(ctx, transaction.ID, err)
		return RegisterResult{}, err
	}

	authCommitted, err := s.commitRegistration(ctx, transaction.ID, credentialID)
	if err != nil {
		return RegisterResult{}, err
	}
	if authCommitted && s.workspaceAdapter != nil {
		// Workspace provisioning is idempotent. Registration has already reached
		// its durable commit decision, so a temporarily unavailable workspace
		// service must not turn that commit into an HTTP failure.
		_ = s.workspaceAdapter.EnsurePersonal(context.WithoutCancel(ctx), transaction.UserID, transaction.UserID, spec.Profile.DisplayName)
	}
	return s.registrationResult(ctx, transaction.UserID, spec.RememberMe, authCommitted), nil
}

func (s *Service) RecoverRegistrations(ctx context.Context, batchSize int) error {
	transactions, err := s.pendingRegistrations(ctx, batchSize)
	if err != nil {
		return err
	}
	var result error
	for _, transaction := range transactions {
		transaction, recoverErr := s.prepareRegistrationRecovery(ctx, transaction.ID)
		if recoverErr != nil {
			result = errors.Join(result, recoverErr)
			continue
		}
		if transaction.terminal() {
			continue
		}
		_, resolveErr := s.resolveRegistration(ctx, transaction)
		result = errors.Join(result, resolveErr)
	}
	return result
}

func (s *Service) startRegistration(ctx context.Context, spec RegisterSpec) (RegistrationTransaction, error) {
	transaction := newRegistrationTransaction(spec.Login, time.Now().UTC())
	err := s.unit.ExecuteWithStorageProvider(ctx, func(provider StorageProvider) error {
		return provider.Registration(ctx).Save(transaction)
	})
	return transaction, err
}

func (s *Service) prepareRegistration(
	ctx context.Context,
	transaction RegistrationTransaction,
	spec RegisterSpec,
) (uuid.UUID, error) {
	var credentialID uuid.UUID
	prepareErrors := fanOut(
		func() error {
			var prepareErr error
			credentialID, prepareErr = s.authAdapter.PrepareRegistration(
				ctx,
				transaction.ID,
				transaction.UserID,
				spec.Login,
				spec.Password,
			)
			return prepareErr
		},
		func() error {
			return s.userAdapter.PrepareRegistration(ctx, transaction.ID, transaction.UserID, spec.Profile)
		},
		func() error {
			if s.accessAdapter == nil {
				return nil
			}
			return s.accessAdapter.PrepareRegistration(ctx, transaction.ID, transaction.UserID)
		},
	)
	return credentialID, errors.Join(prepareErrors...)
}

func (s *Service) abortRegistration(ctx context.Context, transactionID uuid.UUID, cause error) {
	cleanupCtx := context.WithoutCancel(ctx)
	transaction, err := s.updateRegistration(cleanupCtx, transactionID, func(transaction *RegistrationTransaction) error {
		return transaction.decideAbort(cause)
	})
	if err != nil {
		return
	}
	_, _ = s.resolveRegistration(cleanupCtx, transaction)
}

func (s *Service) commitRegistration(ctx context.Context, transactionID, credentialID uuid.UUID) (bool, error) {
	transaction, err := s.updateRegistration(ctx, transactionID, func(transaction *RegistrationTransaction) error {
		return transaction.decideCommit(credentialID)
	})
	if err != nil {
		return false, err
	}
	authCommitted, _ := s.resolveRegistration(context.WithoutCancel(ctx), transaction)
	return authCommitted, nil
}

func (s *Service) registrationResult(
	ctx context.Context,
	userID uuid.UUID,
	rememberMe bool,
	authCommitted bool,
) RegisterResult {
	result := RegisterResult{UserID: userID}
	if !authCommitted {
		return result
	}
	var err error
	result.SessionID, result.ExpiresAt, err = s.sessionAdapter.CreateSession(ctx, userID, rememberMe)
	if err != nil {
		return result
	}
	result.Authenticated = true
	result.Persistent = rememberMe
	return result
}

func (s *Service) pendingRegistrations(ctx context.Context, batchSize int) ([]RegistrationTransaction, error) {
	var transactions []RegistrationTransaction
	err := s.unit.ExecuteWithStorageProvider(ctx, func(provider StorageProvider) error {
		var loadErr error
		transactions, loadErr = provider.Registration(ctx).FindPending(time.Now().UTC(), batchSize)
		return loadErr
	})
	return transactions, err
}

func (s *Service) resolveRegistration(
	ctx context.Context,
	transaction RegistrationTransaction,
) (authResolved bool, result error) {
	commit, err := transaction.decision()
	if err != nil {
		return false, err
	}

	branches := s.registrationBranches(ctx, transaction, commit)
	resolveBranches(branches)
	authResolved = transaction.AuthDone || branches[0].err == nil
	result = errors.Join(branchErrors(branches), s.storeResolution(ctx, transaction.ID, branches))
	return authResolved, result
}

func (s *Service) registrationBranches(
	ctx context.Context,
	transaction RegistrationTransaction,
	commit bool,
) []*branchResolution {
	return []*branchResolution{
		{branch: authBranch, done: transaction.AuthDone, action: func() error {
			if commit {
				return s.authAdapter.CommitRegistration(ctx, transaction.ID)
			}
			return s.authAdapter.RollbackRegistration(ctx, transaction.ID)
		}},
		{branch: userBranch, done: transaction.UserDone, action: func() error {
			if commit {
				return s.userAdapter.CommitRegistration(ctx, transaction.ID)
			}
			return s.userAdapter.RollbackRegistration(ctx, transaction.ID)
		}},
		{branch: accessBranch, done: transaction.AccessDone, action: func() error {
			if s.accessAdapter == nil {
				return nil
			}
			if commit {
				return s.accessAdapter.CommitRegistration(ctx, transaction.ID)
			}
			return s.accessAdapter.RollbackRegistration(ctx, transaction.ID)
		}},
	}
}

func (s *Service) storeResolution(
	ctx context.Context,
	transactionID uuid.UUID,
	branches []*branchResolution,
) error {
	return s.unit.ExecuteWithStorageProvider(ctx, func(provider StorageProvider) error {
		storage := provider.Registration(ctx)
		transaction, err := storage.FindForUpdate(transactionID)
		if err != nil {
			return err
		}
		if transaction.terminal() {
			return nil
		}
		for _, branch := range branches {
			if branch.done || branch.err != nil {
				continue
			}
			if err = transaction.resolveBranch(branch.branch); err != nil {
				return err
			}
		}
		publish, err := transaction.complete()
		if err != nil {
			return err
		}
		if err = storage.Save(transaction); err != nil || !publish {
			return err
		}
		return s.dispatcherFactory.NewDispatcher(ctx).Dispatch(&registererevents.UserRegistered{
			CredentialId: transaction.CredentialID.String(), UserId: transaction.UserID.String(),
			Login: transaction.Login, OccurredAtUnix: time.Now().UTC().Unix(),
		})
	})
}

func (s *Service) prepareRegistrationRecovery(
	ctx context.Context,
	transactionID uuid.UUID,
) (transaction RegistrationTransaction, err error) {
	err = s.unit.ExecuteWithStorageProvider(ctx, func(provider StorageProvider) error {
		storage := provider.Registration(ctx)
		transaction, err = storage.FindForUpdate(transactionID)
		if err != nil || transaction.terminal() {
			return err
		}
		changed, expireErr := transaction.expire(time.Now().UTC())
		if expireErr != nil || !changed {
			return expireErr
		}
		return storage.Save(transaction)
	})
	return transaction, err
}

func (s *Service) updateRegistration(
	ctx context.Context,
	transactionID uuid.UUID,
	update func(*RegistrationTransaction) error,
) (transaction RegistrationTransaction, err error) {
	err = s.unit.ExecuteWithStorageProvider(ctx, func(provider StorageProvider) error {
		storage := provider.Registration(ctx)
		transaction, err = storage.FindForUpdate(transactionID)
		if err != nil {
			return err
		}
		if err = update(&transaction); err != nil {
			return err
		}
		return storage.Save(transaction)
	})
	return transaction, err
}

func resolveBranches(branches []*branchResolution) {
	actions := make([]func() error, 0, len(branches))
	for _, branch := range branches {
		if branch.done {
			continue
		}
		b := branch
		actions = append(actions, func() error {
			b.err = b.action()
			return b.err
		})
	}
	_ = fanOut(actions...)
}

func branchErrors(branches []*branchResolution) (result error) {
	for _, branch := range branches {
		result = errors.Join(result, branch.err)
	}
	return result
}

func fanOut(actions ...func() error) []error {
	errs := make([]error, len(actions))
	var group sync.WaitGroup
	group.Add(len(actions))
	for index, action := range actions {
		i, a := index, action
		go func() {
			defer group.Done()
			errs[i] = a()
		}()
	}
	group.Wait()
	return errs
}
