package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225168 struct{ db mysql.ClientContext }

func Version1791225168(db mysql.ClientContext) migrator.Migration { return &version1791225168{db} }

func (*version1791225168) Version() int64 {
	return 1791225168
}

func (*version1791225168) Description() string {
	return "Create notifications and preferences"
}

func (v *version1791225168) Up(ctx context.Context) error {
	const (
		createNotificationQuery = `
			CREATE TABLE IF NOT EXISTS notification(
			    id BINARY(16) NOT NULL PRIMARY KEY,
			    user_id BINARY(16) NOT NULL,
			    workspace_id BINARY(16) NULL,
			    idempotency_key VARCHAR(191) NOT NULL,
			    kind VARCHAR(80) NOT NULL,
			    title VARCHAR(255) NOT NULL,
			    body TEXT NOT NULL,
			    is_read BOOLEAN NOT NULL DEFAULT FALSE,
			    email_requested BOOLEAN NOT NULL DEFAULT FALSE,
			    recipient_email VARCHAR(320) NULL,
			    created_at DATETIME(6) NOT NULL,
			    UNIQUE KEY notification_idempotency_uidx(user_id, idempotency_key),
			    INDEX notification_user_created_idx(user_id, created_at)
			)
		`
		createNotificationPreferencesQuery = `
			CREATE TABLE IF NOT EXISTS notification_preferences(
			    user_id BINARY(16) NOT NULL PRIMARY KEY,
			    in_app BOOLEAN NOT NULL,
			    email BOOLEAN NOT NULL
			)
		`
	)
	statements := []string{createNotificationQuery, createNotificationPreferencesQuery}

	for i, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := v.db.ExecContext(ctx, fmt.Sprintf("notification.schema.%d", i), statement); err != nil {
			return err
		}
	}
	return nil
}
