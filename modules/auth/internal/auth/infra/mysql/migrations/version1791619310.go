package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791619310 struct {
	client mysql.ClientContext
}

func Version1791619310(client mysql.ClientContext) migrator.Migration {
	return &version1791619310{
		client: client,
	}
}

func (v *version1791619310) Version() int64 {
	return 1791619310
}

func (v *version1791619310) Description() string {
	return "Create credentials table"
}

func (v *version1791619310) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-credential-table", `
		CREATE TABLE IF NOT EXISTS credential (
			id BINARY(16) NOT NULL PRIMARY KEY,
			user_id BINARY(16) NOT NULL,
			login VARCHAR(255) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			method VARCHAR(32) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			deleted_at DATETIME(6) NULL,
			UNIQUE KEY credential_login_uidx(login),
			INDEX credential_user_idx(user_id)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
