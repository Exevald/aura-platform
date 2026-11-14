package repository

import (
	"context"
	"database/sql"
	"errors"
	"notification/internal/notification/domain"
	"time"
	"uuid"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type repo struct {
	ctx context.Context
	db  sdkmysql.ClientContext
}

func New(ctx context.Context, db sdkmysql.ClientContext) domain.Repository {
	return &repo{ctx: ctx, db: db}
}

func (r *repo) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *repo) Store(v domain.Notification) (bool, error) {
	const sqlNotificationStoreQuery = `
		INSERT IGNORE INTO notification(
		    id,
		    user_id,
		    workspace_id,
		    idempotency_key,
		    kind,
		    title,
		    body,
		    is_read,
		    email_requested,
		    recipient_email,
		    created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var wid any
	if v.WorkspaceID != uuid.Nil() {
		wid = v.WorkspaceID
	}
	var recipient any
	if v.RecipientEmail != "" {
		recipient = v.RecipientEmail
	}
	res, e := r.db.ExecContext(r.ctx, "notification.store", sqlNotificationStoreQuery, v.ID, v.UserID, wid, v.IdempotencyKey, v.Kind, v.Title, v.Body, v.Read, v.EmailRequested, recipient, v.CreatedAt)
	if e != nil {
		return false, e
	}
	n, e := res.RowsAffected()
	return n > 0, e
}

func (r *repo) List(user uuid.UUID) ([]domain.Notification, error) {
	const sqlNotificationListQuery = `
		SELECT
		    id,
		    user_id,
		    workspace_id,
		    idempotency_key,
		    kind,
		    title,
		    body,
		    is_read,
		    email_requested,
		    recipient_email,
		    created_at
		FROM notification
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT 200
	`

	var rows []row
	e := r.db.SelectContext(r.ctx, "notification.list", &rows, sqlNotificationListQuery, user)
	out := make([]domain.Notification, 0, len(rows))
	for _, v := range rows {
		n := domain.Notification{ID: v.ID, UserID: v.UserID, IdempotencyKey: v.IdempotencyKey, Kind: v.Kind, Title: v.Title, Body: v.Body, Read: v.Read, EmailRequested: v.EmailRequested, RecipientEmail: v.RecipientEmail.String, CreatedAt: v.CreatedAt}
		if v.WorkspaceID != nil {
			n.WorkspaceID = *v.WorkspaceID
		}
		out = append(out, n)
	}
	return out, e
}

func (r *repo) MarkRead(user, id uuid.UUID) error {
	const sqlNotificationReadQuery = `
		UPDATE notification
		SET
		    is_read = TRUE
		WHERE id = ? AND user_id = ?
	`

	res, e := r.db.ExecContext(r.ctx, "notification.read", sqlNotificationReadQuery, id, user)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

func (r *repo) GetPreferences(user uuid.UUID) (domain.Preferences, error) {
	const sqlNotificationPreferencesGetQuery = `
		SELECT
		    user_id,
		    in_app,
		    email
		FROM notification_preferences
		WHERE user_id = ?
	`

	var v struct {
		UserID uuid.UUID `db:"user_id"`
		InApp  bool      `db:"in_app"`
		Email  bool      `db:"email"`
	}
	e := r.db.GetContext(r.ctx, "notification.preferences.get", &v, sqlNotificationPreferencesGetQuery, user)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return domain.Preferences{UserID: v.UserID, InApp: v.InApp, Email: v.Email}, e
}

func (r *repo) StorePreferences(v domain.Preferences) error {
	const sqlNotificationPreferencesStoreQuery = `
		INSERT INTO notification_preferences(
		    user_id,
		    in_app,
		    email
		)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    in_app = VALUES (in_app),
		    email = VALUES (email)
	`

	_, e := r.db.ExecContext(r.ctx, "notification.preferences.store", sqlNotificationPreferencesStoreQuery, v.UserID, v.InApp, v.Email)
	return e
}

type row struct {
	ID             uuid.UUID      `db:"id"`
	UserID         uuid.UUID      `db:"user_id"`
	WorkspaceID    *uuid.UUID     `db:"workspace_id"`
	IdempotencyKey string         `db:"idempotency_key"`
	Kind           string         `db:"kind"`
	Title          string         `db:"title"`
	Body           string         `db:"body"`
	Read           bool           `db:"is_read"`
	EmailRequested bool           `db:"email_requested"`
	RecipientEmail sql.NullString `db:"recipient_email"`
	CreatedAt      time.Time      `db:"created_at"`
}
