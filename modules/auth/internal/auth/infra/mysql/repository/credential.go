package repository

import (
	"context"
	"database/sql"
	stderrors "errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"auth/internal/auth/domain"
)

type credentialRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func NewCredentialRepository(
	ctx context.Context,
	client mysql.ClientContext,
) domain.CredentialRepository {
	return &credentialRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *credentialRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *credentialRepository) Store(credential domain.Credential) error {
	const query = `
		INSERT INTO credential (
			id,
			user_id,
			login,
			password_hash,
			method,
			created_at,
			updated_at,
			deleted_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, NULL)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			login = VALUES(login),
			password_hash = VALUES(password_hash),
			method = VALUES(method),
			updated_at = VALUES(updated_at),
			deleted_at = NULL
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"store-credential",
		query,
		newBinaryUUID(credential.ID),
		newBinaryUUID(credential.UserID),
		credential.Login,
		credential.PasswordHash,
		credential.Method,
		credential.CreatedAt,
		credential.UpdatedAt,
	)
	var mysqlErr *mysqldriver.MySQLError
	if stderrors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return errors.WithStack(domain.ErrCredentialExists)
	}
	return errors.WithStack(err)
}

func (r *credentialRepository) Find(id uuid.UUID) (domain.Credential, error) {
	const query = `
		SELECT 
		    id,
		    user_id,
		    login,
		    password_hash,
		    method,
		    created_at,
		    updated_at
		FROM credential
		WHERE id = ? AND deleted_at IS NULL
	`
	return r.find("find-credential", query, newBinaryUUID(id))
}

func (r *credentialRepository) FindByLogin(login string) (domain.Credential, error) {
	const query = `
		SELECT 
		    id,
		    user_id,
		    login,
		    password_hash,
		    method,
		    created_at,
		    updated_at
		FROM credential
		WHERE login = ? AND deleted_at IS NULL
	`
	return r.find("find-credential-by-login", query, login)
}

func (r *credentialRepository) FindByUserID(userID uuid.UUID) (domain.Credential, error) {
	const query = `
		SELECT 
		    id,
		    user_id,
		    login,
		    password_hash,
		    method,
		    created_at,
		    updated_at
		FROM credential
		WHERE user_id = ? AND deleted_at IS NULL
	`
	return r.find("find-credential-by-user-id", query, newBinaryUUID(userID))
}

func (r *credentialRepository) find(name, query string, arg any) (domain.Credential, error) {
	var row credentialRow
	err := r.client.GetContext(r.ctx, name, &row, query, arg)
	if stderrors.Is(err, sql.ErrNoRows) {
		return domain.Credential{}, errors.WithStack(domain.ErrCredentialNotFound)
	}
	if err != nil {
		return domain.Credential{}, errors.WithStack(err)
	}

	return row.toDomainModel(), nil
}

func (r *credentialRepository) Delete(id uuid.UUID) error {
	const query = `UPDATE credential SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL`
	_, err := r.client.ExecContext(r.ctx, "delete-credential", query, newBinaryUUID(id))
	return errors.WithStack(err)
}

type credentialRow struct {
	ID           binaryUUID        `db:"id"`
	UserID       binaryUUID        `db:"user_id"`
	Login        string            `db:"login"`
	PasswordHash string            `db:"password_hash"`
	Method       domain.AuthMethod `db:"method"`
	CreatedAt    time.Time         `db:"created_at"`
	UpdatedAt    time.Time         `db:"updated_at"`
}

func (r *credentialRow) toDomainModel() domain.Credential {
	return domain.Credential{
		ID:           r.ID.UUID(),
		UserID:       r.UserID.UUID(),
		Login:        r.Login,
		PasswordHash: r.PasswordHash,
		Method:       r.Method,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
