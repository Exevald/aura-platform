package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791619307 struct {
	client mysql.ClientContext
}

func Version1791619307(client mysql.ClientContext) migrator.Migration {
	return &version1791619307{
		client: client,
	}
}

func (v *version1791619307) Version() int64 {
	return 1791619307
}

func (v *version1791619307) Description() string {
	return "Create module access tables"
}

func (v *version1791619307) Up(ctx context.Context) error {
	if _, err := v.client.ExecContext(ctx, "create-module-access-table", `
		CREATE TABLE IF NOT EXISTS module_access (
			id BINARY(16) NOT NULL PRIMARY KEY,
			user_id BINARY(16) NOT NULL,
			module_id BINARY(16) NOT NULL,
			effect VARCHAR(16) NOT NULL,
			source_id VARCHAR(64) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			UNIQUE KEY module_access_source_user_module_uidx(source_id, user_id, module_id),
			INDEX module_access_user_module_idx(user_id, module_id)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}
	_, err := v.client.ExecContext(ctx, "create-module-share-operation-table", `
		CREATE TABLE IF NOT EXISTS module_share_operation (
			share_id BINARY(16) NOT NULL PRIMARY KEY,
			module_id BINARY(16) NOT NULL,
			group_id BINARY(16) NOT NULL,
			owner_user_id BINARY(16) NOT NULL,
			status ENUM('granted','cancelled') NOT NULL,
			created_at DATETIME(6) NOT NULL,
			updated_at DATETIME(6) NOT NULL,
			INDEX module_share_operation_group_idx(group_id, status)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
