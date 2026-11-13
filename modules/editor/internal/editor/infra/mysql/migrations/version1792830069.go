package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1792830069 struct {
	client mysql.ClientContext
}

func Version1792830069(client mysql.ClientContext) migrator.Migration {
	return &version1792830069{
		client: client,
	}
}

func (v *version1792830069) Version() int64 {
	return 1792830069
}

func (v *version1792830069) Description() string {
	return "Create a document_permission table"
}

func (v *version1792830069) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-document-permission-table", `
		CREATE TABLE IF NOT EXISTS document_permission (
		 document_id BINARY(16) NOT NULL,
		 user_id BINARY(16) NOT NULL,
		 role ENUM('viewer','editor') NOT NULL,
		 created_at DATETIME(6) NOT NULL,
		 updated_at DATETIME(6) NOT NULL,
		 PRIMARY KEY (document_id, user_id),
		 INDEX document_permission_user_idx (user_id),
		 CONSTRAINT document_permission_document_fk FOREIGN KEY (document_id) REFERENCES document(id) ON DELETE CASCADE
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`)
	return err
}
