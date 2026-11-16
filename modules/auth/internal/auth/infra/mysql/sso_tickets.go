package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"uuid"

	clientmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"auth/internal/auth/app"
)

type SSOTicketStore struct {
	client clientmysql.TransactionalClient
}

func NewSSOTicketStore(client clientmysql.TransactionalClient) *SSOTicketStore {
	return &SSOTicketStore{client: client}
}

func (s *SSOTicketStore) Save(ctx context.Context, hash []byte, userID uuid.UUID, target string, expiresAt time.Time) error {
	if _, err := s.client.ExecContext(ctx, "expire-sso-tickets", "DELETE FROM sso_ticket WHERE expires_at <= UTC_TIMESTAMP(6)"); err != nil {
		return err
	}
	_, err := s.client.ExecContext(ctx, "save-sso-ticket", `
		INSERT INTO sso_ticket (code_hash, user_id, target_host, expires_at)
		VALUES (?, ?, ?, ?)`, hash, userID.String(), target, expiresAt)
	return err
}

func (s *SSOTicketStore) Take(ctx context.Context, hash []byte, target string) (_ uuid.UUID, err error) {
	tx, err := s.client.BeginTransaction()
	if err != nil {
		return uuid.Nil(), err
	}
	defer func() { _ = tx.Rollback() }()
	var userID string
	err = tx.QueryRowContext(ctx, "lock-sso-ticket", `
		SELECT user_id FROM sso_ticket
		WHERE code_hash = ? AND target_host = ? AND expires_at > UTC_TIMESTAMP(6)
		FOR UPDATE`, hash, target).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil(), app.ErrInvalidSSOTicket
	}
	if err != nil {
		return uuid.Nil(), err
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil(), err
	}
	if _, err = tx.ExecContext(ctx, "delete-sso-ticket", "DELETE FROM sso_ticket WHERE code_hash = ?", hash); err != nil {
		return uuid.Nil(), err
	}
	if err = tx.Commit(); err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}
