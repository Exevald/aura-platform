package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791619308 struct {
	client mysql.ClientContext
}

func Version1791619308(client mysql.ClientContext) migrator.Migration {
	return &version1791619308{
		client: client,
	}
}

func (v *version1791619308) Version() int64 {
	return 1791619308
}

func (v *version1791619308) Description() string {
	return "Create module share operation table"
}

func (v *version1791619308) Up(ctx context.Context) error {
	const sqlCreateModuleShareOperationTableQuery = `
			CREATE TABLE IF NOT EXISTS module_share_operation (
			    share_id BINARY(16) NOT NULL PRIMARY KEY,
			    module_id BINARY(16) NOT NULL,
			    group_id BINARY(16) NOT NULL,
			    owner_user_id BINARY(16) NOT NULL,
			    status ENUM('granted', 'cancelled') NOT NULL,
			    created_at DATETIME(6) NOT NULL,
			    updated_at DATETIME(6) NOT NULL,
			    active_group_module_key VARCHAR(96)
			      GENERATED ALWAYS AS (
			        CASE WHEN status = 'granted'
			          THEN CONCAT(HEX(group_id), HEX(owner_user_id), HEX(module_id))
			          ELSE NULL END
			      ) STORED,
			    INDEX module_share_operation_group_idx(group_id, status),
			    UNIQUE KEY module_share_active_group_module_uidx(active_group_module_key)
			) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
		`

	_, err := v.client.ExecContext(ctx, "create-module-share-operation-table", sqlCreateModuleShareOperationTableQuery)
	return err
}
