package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1793000010 struct{ client mysql.ClientContext }

func Version1793000010(client mysql.ClientContext) migrator.Migration {
	return &version1793000010{client: client}
}

func (*version1793000010) Version() int64 {
	return 1793000010
}

func (*version1793000010) Description() string {
	return "Create module share saga"
}

func (m *version1793000010) Up(ctx context.Context) error {
	const sqlCreateModuleShareQuery = `
		CREATE TABLE IF NOT EXISTS module_share (
		    id BINARY(16) NOT NULL PRIMARY KEY,
		    module_id BINARY(16) NOT NULL,
		    group_id BINARY(16) NOT NULL,
		    owner_user_id BINARY(16) NOT NULL,
		    status ENUM('pending', 'active', 'compensating', 'failed', 'revoked') NOT NULL,
		    attempt INT NOT NULL,
		    failure_reason VARCHAR(255) NULL,
		    created_at DATETIME(6) NOT NULL,
		    updated_at DATETIME(6) NOT NULL,
		    INDEX module_share_owner_idx(owner_user_id, created_at),
		    INDEX module_share_pending_idx(status, updated_at)
		) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
	`

	_, err := m.client.ExecContext(ctx, "create-module-share", sqlCreateModuleShareQuery)
	return err
}
