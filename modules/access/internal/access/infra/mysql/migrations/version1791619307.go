package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1791619307(client mysql.ClientContext) migrator.Migration {
	return &version1791619307{
		client: client,
	}
}

type version1791619307 struct {
	client mysql.ClientContext
}

func (v *version1791619307) Version() int64 {
	return 1791619307
}

func (v *version1791619307) Description() string {
	return "Create module access table"
}

func (v *version1791619307) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-module-access-table", `
		CREATE TABLE IF NOT EXISTS module_access (
			id BINARY(16) NOT NULL PRIMARY KEY,
			user_id BINARY(16) NOT NULL,
			module_id BINARY(16) NOT NULL,
			effect VARCHAR(16) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			INDEX module_access_user_module_idx(user_id, module_id)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
