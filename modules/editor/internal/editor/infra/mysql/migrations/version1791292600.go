package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1791292600(client mysql.ClientContext) migrator.Migration {
	return &documentMigration{client: client}
}

type documentMigration struct{ client mysql.ClientContext }

func (*documentMigration) Version() int64 {
	return 1791292600
}

func (*documentMigration) Description() string {
	return "Create editor documents and permissions"
}

func (m *documentMigration) Up(ctx context.Context) error {
	_, err := m.client.ExecContext(ctx, "create-document-table", `
		CREATE TABLE IF NOT EXISTS document (
		 id BINARY(16) NOT NULL PRIMARY KEY,
		 owner_user_id BINARY(16) NOT NULL,
		 title VARCHAR(255) NOT NULL,
		 content MEDIUMTEXT NOT NULL,
		 version BIGINT NOT NULL,
		 created_at DATETIME(6) NOT NULL,
		 updated_at DATETIME(6) NOT NULL,
		 INDEX document_owner_idx (owner_user_id)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
