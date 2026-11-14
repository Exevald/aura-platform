package storage

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"registerer/internal/registerer/app"
)

type registrationStorage struct {
	ctx    context.Context
	client mysql.ClientContext
}

type registrationRow struct {
	ID           binaryUUID            `db:"id"`
	UserID       binaryUUID            `db:"user_id"`
	CredentialID []byte                `db:"credential_id"`
	Login        string                `db:"login"`
	State        app.RegistrationState `db:"state"`
	AuthDone     bool                  `db:"auth_done"`
	UserDone     bool                  `db:"user_done"`
	AccessDone   bool                  `db:"access_done"`
	LeaseUntil   *time.Time            `db:"lease_until"`
	LastError    string                `db:"last_error"`
}

func NewRegistrationStorage(ctx context.Context, client mysql.ClientContext) app.RegistrationStorage {
	return &registrationStorage{ctx: ctx, client: client}
}

func (s *registrationStorage) Save(transaction app.RegistrationTransaction) error {
	var credentialID any
	if transaction.CredentialID != uuid.Nil() {
		credentialID = newBinaryUUID(transaction.CredentialID)
	}
	var leaseUntil any
	if !transaction.LeaseUntil.IsZero() {
		leaseUntil = transaction.LeaseUntil
	}
	const query = `
		INSERT INTO registration_transaction (
		    id,
		    user_id,
		    credential_id,
		    login,
		    state,
		    auth_done,
		    user_done,
		    access_done,
		    lease_until,
		    last_error,
		    created_at,
		    updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), UTC_TIMESTAMP(6), UTC_TIMESTAMP(6))
		ON DUPLICATE KEY UPDATE
		    user_id = VALUES (user_id),
		    credential_id = VALUES (credential_id),
		    login = VALUES (login),
		    state = VALUES (state),
		    auth_done = VALUES (auth_done),
		    user_done = VALUES (user_done),
		    access_done = VALUES (access_done),
		    lease_until = VALUES (lease_until),
		    last_error = VALUES (last_error),
		    updated_at = UTC_TIMESTAMP(6)
	`
	_, err := s.client.ExecContext(
		s.ctx,
		"registration.save",
		query,
		newBinaryUUID(transaction.ID),
		newBinaryUUID(transaction.UserID),
		credentialID,
		transaction.Login,
		transaction.State,
		transaction.AuthDone,
		transaction.UserDone,
		transaction.AccessDone,
		leaseUntil,
		transaction.LastError,
	)
	return err
}

func (s *registrationStorage) FindForUpdate(id uuid.UUID) (app.RegistrationTransaction, error) {
	const sqlRegistrationFindForUpdateQuery = `
		SELECT
		    id,
		    user_id,
		    credential_id,
		    login,
		    state,
		    auth_done,
		    user_done,
		    access_done,
		    lease_until,
		    COALESCE(last_error, '') AS last_error
		FROM registration_transaction
		WHERE id = ?
		FOR UPDATE
	`

	var row registrationRow
	err := s.client.GetContext(s.ctx, "registration.find-for-update", &row, sqlRegistrationFindForUpdateQuery, newBinaryUUID(id))
	return row.registration(), err
}

func (s *registrationStorage) FindPending(now time.Time, limit int) ([]app.RegistrationTransaction, error) {
	const sqlRegistrationPendingQuery = `
		SELECT
		    id,
		    user_id,
		    credential_id,
		    login,
		    state,
		    auth_done,
		    user_done,
		    access_done,
		    lease_until,
		    COALESCE(last_error, '') AS last_error
		FROM registration_transaction
		WHERE state IN ('committing', 'aborting') OR (state = 'preparing' AND lease_until < ?)
		ORDER BY updated_at
		LIMIT ?
	`

	var rows []registrationRow
	err := s.client.SelectContext(s.ctx, "registration.pending", &rows, sqlRegistrationPendingQuery, now.UTC(), limit)
	if err != nil {
		return nil, err
	}
	result := make([]app.RegistrationTransaction, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.registration())
	}
	return result, nil
}

func (r registrationRow) registration() app.RegistrationTransaction {
	transaction := app.RegistrationTransaction{
		ID: r.ID.UUID(), UserID: r.UserID.UUID(), Login: r.Login, State: r.State,
		AuthDone: r.AuthDone, UserDone: r.UserDone, AccessDone: r.AccessDone, LastError: r.LastError,
	}
	if len(r.CredentialID) == 16 {
		copy(transaction.CredentialID[:], r.CredentialID)
	}
	if r.LeaseUntil != nil {
		transaction.LeaseUntil = *r.LeaseUntil
	}
	return transaction
}
