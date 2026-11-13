package app

import (
	"errors"
	"testing"
	"time"
	"uuid"
)

func TestRegistrationTransactionCommitDecision(t *testing.T) {
	credentialID := uuid.New()
	transaction := preparingTransaction()

	if err := transaction.decideCommit(uuid.Nil()); !errors.Is(err, errInvalidRegistrationTransition) {
		t.Fatalf("decideCommit(nil) error = %v, want invalid transition", err)
	}
	if err := transaction.decideCommit(credentialID); err != nil {
		t.Fatalf("decideCommit() error = %v", err)
	}
	if transaction.State != RegistrationCommitting {
		t.Fatalf("state = %q, want %q", transaction.State, RegistrationCommitting)
	}
	if transaction.CredentialID != credentialID {
		t.Fatalf("credential id = %s, want %s", transaction.CredentialID, credentialID)
	}
	if !transaction.LeaseUntil.IsZero() || transaction.LastError != "" {
		t.Fatal("commit decision did not clear lease and error")
	}
	if err := transaction.decideAbort(errors.New("late failure")); !errors.Is(err, errInvalidRegistrationTransition) {
		t.Fatalf("decideAbort() after commit error = %v, want invalid transition", err)
	}
}

func TestRegistrationTransactionExpiresToAbort(t *testing.T) {
	now := time.Now().UTC()
	transaction := preparingTransaction()
	transaction.LeaseUntil = now.Add(time.Minute)

	changed, err := transaction.expire(now)
	if err != nil || changed {
		t.Fatalf("expire() before lease = (%v, %v), want (false, nil)", changed, err)
	}
	transaction.LeaseUntil = now.Add(-time.Nanosecond)
	changed, err = transaction.expire(now)
	if err != nil || !changed {
		t.Fatalf("expire() after lease = (%v, %v), want (true, nil)", changed, err)
	}
	if transaction.State != RegistrationAborting || transaction.LastError != expiredRegistrationError {
		t.Fatalf("expired transaction = state %q, error %q", transaction.State, transaction.LastError)
	}
	if err = transaction.decideAbort(errors.New("duplicate")); err != nil {
		t.Fatalf("idempotent decideAbort() error = %v", err)
	}
}

func TestRegistrationTransactionCompletesCommitOnce(t *testing.T) {
	transaction := preparingTransaction()
	if err := transaction.decideCommit(uuid.New()); err != nil {
		t.Fatalf("decideCommit() error = %v", err)
	}

	publish, err := transaction.complete()
	if err != nil || publish || transaction.State != RegistrationCommitting {
		t.Fatalf("incomplete complete() = (%v, %v, %q)", publish, err, transaction.State)
	}
	for _, branch := range []registrationBranch{authBranch, userBranch, accessBranch} {
		if err = transaction.resolveBranch(branch); err != nil {
			t.Fatalf("resolveBranch(%q) error = %v", branch, err)
		}
	}
	publish, err = transaction.complete()
	if err != nil || !publish || transaction.State != RegistrationCommitted {
		t.Fatalf("complete() = (%v, %v, %q), want published commit", publish, err, transaction.State)
	}
	publish, err = transaction.complete()
	if err != nil || publish {
		t.Fatalf("second complete() = (%v, %v), want (false, nil)", publish, err)
	}
}

func TestRegistrationTransactionCompletesAbortWithoutEvent(t *testing.T) {
	transaction := preparingTransaction()
	if err := transaction.decideAbort(errors.New("prepare failed")); err != nil {
		t.Fatalf("decideAbort() error = %v", err)
	}
	for _, branch := range []registrationBranch{authBranch, userBranch, accessBranch} {
		if err := transaction.resolveBranch(branch); err != nil {
			t.Fatalf("resolveBranch(%q) error = %v", branch, err)
		}
	}
	publish, err := transaction.complete()
	if err != nil || publish || transaction.State != RegistrationAborted {
		t.Fatalf("complete() = (%v, %v, %q), want silent abort", publish, err, transaction.State)
	}
}

func preparingTransaction() RegistrationTransaction {
	return newRegistrationTransaction("login", time.Now().UTC())
}
