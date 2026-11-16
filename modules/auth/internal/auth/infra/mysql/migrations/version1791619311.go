package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791619311 struct{ client mysql.ClientContext }

func Version1791619311(client mysql.ClientContext) migrator.Migration {
	return &version1791619311{client: client}
}

func (v *version1791619311) Version() int64 {
	return 1791619311
}

func (v *version1791619311) Description() string {
	return "Create one-time SSO tickets table"
}

func (v *version1791619311) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-sso-ticket-table", `
		CREATE TABLE IF NOT EXISTS sso_ticket (
			code_hash BINARY(32) NOT NULL PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			target_host VARCHAR(128) NOT NULL,
			expires_at DATETIME(6) NOT NULL,
			INDEX sso_ticket_expiry_idx (expires_at)
		) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci`)
	return err
}
