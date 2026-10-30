package repository

import (
	"context"
	"database/sql"
	stderrors "errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"

	"session/internal/session/domain"
)

func NewSessionRepository(
	ctx context.Context,
	client mysql.ClientContext,
) domain.SessionRepository {
	return &sessionRepository{
		ctx:    ctx,
		client: client,
	}
}

type sessionRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (r *sessionRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *sessionRepository) Store(v domain.Session) error {
	const query = `
		INSERT INTO session (id, user_id, status, created_at, expires_at, revoked_at, persistent)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			expires_at = VALUES(expires_at),
			revoked_at = VALUES(revoked_at),
			persistent = VALUES(persistent)
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"store-session",
		query,
		v.ID,
		v.UserID,
		v.Status,
		v.CreatedAt,
		v.ExpiresAt,
		v.RevokedAt,
		v.Persistent,
	)
	return errors.WithStack(err)
}

func (r *sessionRepository) Find(id uuid.UUID) (domain.Session, error) {
	const query = `
		SELECT
			id,
			user_id,
			status,
			created_at,
			expires_at,
			revoked_at
			,persistent
		FROM session
		WHERE id = ?
	`
	var row sessionRow
	err := r.client.GetContext(
		r.ctx,
		"find-session",
		&row,
		query,
		id,
	)
	if stderrors.Is(err, sql.ErrNoRows) {
		return domain.Session{}, errors.WithStack(domain.ErrSessionNotFound)
	}
	if err != nil {
		return domain.Session{}, errors.WithStack(err)
	}

	return row.toDomainModel(), nil
}

func (r *sessionRepository) FindActiveByUser(userID uuid.UUID) ([]domain.Session, error) {
	const query = `
		SELECT
			id,
			user_id,
			status,
			created_at,
			expires_at,
			revoked_at
			,persistent
		FROM session
		WHERE user_id = ? AND status = 'active'
	`
	var rows []sessionRow
	err := r.client.SelectContext(
		r.ctx,
		"find-active-sessions-by-user",
		&rows,
		query,
		userID,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	out := make([]domain.Session, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomainModel())
	}
	return out, nil
}

type sessionRow struct {
	ID         uuid.UUID            `db:"id"`
	UserID     uuid.UUID            `db:"user_id"`
	Status     domain.SessionStatus `db:"status"`
	CreatedAt  time.Time            `db:"created_at"`
	ExpiresAt  time.Time            `db:"expires_at"`
	RevokedAt  *time.Time           `db:"revoked_at"`
	Persistent bool                 `db:"persistent"`
}

func (r *sessionRow) toDomainModel() domain.Session {
	return domain.Session{
		ID:         r.ID,
		UserID:     r.UserID,
		Status:     r.Status,
		CreatedAt:  r.CreatedAt,
		ExpiresAt:  r.ExpiresAt,
		RevokedAt:  r.RevokedAt,
		Persistent: r.Persistent,
	}
}
