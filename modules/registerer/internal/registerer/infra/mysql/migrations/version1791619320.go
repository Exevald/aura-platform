package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791619320 struct {
	client mysql.ClientContext
}

func Version1791619320(client mysql.ClientContext) migrator.Migration {
	return &version1791619320{client: client}
}

func (*version1791619320) Version() int64 {
	return 1791619320
}

func (*version1791619320) Description() string {
	return "Create registration coordinator table"
}

func (m *version1791619320) Up(ctx context.Context) error {
	const sqlCreateRegistrationTransactionTableQuery = `
		CREATE TABLE IF NOT EXISTS registration_transaction (
		    id BINARY(16) NOT NULL PRIMARY KEY,
		    user_id BINARY(16) NOT NULL,
		    credential_id BINARY(16) NULL,
		    login VARCHAR(255) NOT NULL,
		    state VARCHAR(16) NOT NULL,
		    auth_done BOOLEAN NOT NULL DEFAULT FALSE,
		    user_done BOOLEAN NOT NULL DEFAULT FALSE,
		    access_done BOOLEAN NOT NULL DEFAULT FALSE,
		    lease_until DATETIME(6) NULL,
		    last_error TEXT NULL,
		    created_at DATETIME(6) NOT NULL,
		    updated_at DATETIME(6) NOT NULL,
		    INDEX registration_pending_idx(state, updated_at)
		) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
	`

	_, err := m.client.ExecContext(ctx, "create-registration-transaction-table", sqlCreateRegistrationTransactionTableQuery)
	return err
}
