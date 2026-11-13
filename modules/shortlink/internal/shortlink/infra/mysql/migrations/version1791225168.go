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

func (*version1791225168) Version() int64 {
	return 1791225168
}

func (*version1791225168) Description() string {
	return "Create short links"
}

func (v *version1791225168) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-short-link", `
		CREATE TABLE IF NOT EXISTS short_link (
			id CHAR(36) NOT NULL PRIMARY KEY,
			owner_user_id CHAR(36) NOT NULL,
			code CHAR(10) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
			target_url VARCHAR(2048) NOT NULL,
			created_at DATETIME(6) NOT NULL,
			UNIQUE KEY short_link_code_uidx(code),
			INDEX short_link_owner_created_idx(owner_user_id, created_at)
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
