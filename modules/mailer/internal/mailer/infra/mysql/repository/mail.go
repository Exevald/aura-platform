package repository

import (
	"context"
	"database/sql"
	"errors"
	"mailer/internal/mailer/domain"
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

type row struct {
	ID             uuid.UUID      `db:"id"`
	IdempotencyKey string         `db:"idempotency_key"`
	Recipient      string         `db:"recipient"`
	Subject        string         `db:"subject"`
	Body           string         `db:"body"`
	Status         string         `db:"status"`
	Attempts       int            `db:"attempts"`
	LastError      sql.NullString `db:"last_error"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func (r *repo) Create(v domain.Job) (domain.Job, bool, error) {
	res, e := r.db.ExecContext(r.ctx, "mail.create", `INSERT IGNORE INTO mail_job(id,idempotency_key,recipient,subject,body,status,attempts,last_error,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, v.ID, v.IdempotencyKey, v.Recipient, v.Subject, v.Body, v.Status, v.Attempts, nil, v.CreatedAt, v.UpdatedAt)
	if e != nil {
		return domain.Job{}, false, e
	}
	n, e := res.RowsAffected()
	if e != nil {
		return domain.Job{}, false, e
	}
	if n == 0 {
		x, e := r.FindByKey(v.IdempotencyKey)
		return x, false, e
	}
	return v, true, nil
}

func (r *repo) Find(id uuid.UUID) (domain.Job, error) {
	return r.find(`SELECT id,idempotency_key,recipient,subject,body,status,attempts,last_error,created_at,updated_at FROM mail_job WHERE id=?`, id)
}

func (r *repo) FindByKey(k string) (domain.Job, error) {
	return r.find(`SELECT id,idempotency_key,recipient,subject,body,status,attempts,last_error,created_at,updated_at FROM mail_job WHERE idempotency_key=?`, k)
}

func (r *repo) find(q string, arg any) (domain.Job, error) {
	var v row
	e := r.db.GetContext(r.ctx, "mail.find", &v, q, arg)
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Job{}, domain.ErrNotFound
	}
	return domain.Job{ID: v.ID, IdempotencyKey: v.IdempotencyKey, Recipient: v.Recipient, Subject: v.Subject, Body: v.Body, Status: v.Status, Attempts: v.Attempts, LastError: v.LastError.String, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}, e
}

func (r *repo) StoreResult(v domain.Job) error {
	_, e := r.db.ExecContext(r.ctx, "mail.result", `UPDATE mail_job SET status=?,attempts=?,last_error=?,updated_at=? WHERE id=?`, v.Status, v.Attempts, nullString(v.LastError), v.UpdatedAt, v.ID)
	return e
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
