package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1791619312(client mysql.ClientContext) migrator.Migration {
	return &version1791619312{
		client: client,
	}
}

type version1791619312 struct {
	client mysql.ClientContext
}

func (v *version1791619312) Version() int64 {
	return 1791619312
}

func (v *version1791619312) Description() string {
	return "Create sessions table"
}

func (v *version1791619312) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-session-table", `
		CREATE TABLE IF NOT EXISTS session (
			id BINARY(16) NOT NULL PRIMARY KEY,
			user_id BINARY(16) NOT NULL,
			status VARCHAR(16) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			expires_at DATETIME(6) NOT NULL,
			revoked_at DATETIME(6) NULL,
		    persistent BOOLEAN NOT NULL DEFAULT FALSE,
			INDEX session_user_status_idx(user_id, status)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
