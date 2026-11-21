package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225168 struct{ client mysql.ClientContext }

func Version1791225168(client mysql.ClientContext) migrator.Migration {
	return &version1791225168{client: client}
}

func (*version1791225168) Version() int64 { return 1791225168 }

func (*version1791225168) Description() string { return "Create publication releases" }

func (v *version1791225168) Up(ctx context.Context) error {
	const query = `CREATE TABLE IF NOT EXISTS publication_release (
			id CHAR(36) PRIMARY KEY,
			owner_user_id CHAR(36) NOT NULL,
			workspace_id CHAR(36) NOT NULL,
			document_id CHAR(36) NOT NULL,
			idempotency_key VARCHAR(255) NOT NULL,
			expected_version BIGINT NOT NULL,
			source_version BIGINT NOT NULL DEFAULT 0,
			title VARCHAR(255) NOT NULL DEFAULT '',
			content MEDIUMTEXT NOT NULL,
			status VARCHAR(16) NOT NULL,
			link_id CHAR(36) NOT NULL DEFAULT '',
			short_path VARCHAR(32) NOT NULL DEFAULT '',
			failure TEXT NOT NULL,
			delivery_status VARCHAR(16) NOT NULL DEFAULT 'pending',
			created_at DATETIME(6) NOT NULL,
			published_at DATETIME(6) NULL,
    		revision BIGINT NOT NULL DEFAULT 1,
			UNIQUE KEY publication_idempotency (owner_user_id, idempotency_key),
			KEY publication_document (document_id),
			KEY publication_status (status)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	_, err := v.client.ExecContext(ctx, "create-publication-release", query)
	return err
}
