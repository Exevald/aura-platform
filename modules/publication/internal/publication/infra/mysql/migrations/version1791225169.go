package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225169 struct{ client mysql.ClientContext }

func Version1791225169(client mysql.ClientContext) migrator.Migration {
	return &version1791225169{client: client}
}

func (*version1791225169) Version() int64 { return 1791225169 }

func (*version1791225169) Description() string { return "Create publication attachments" }

func (v *version1791225169) Up(ctx context.Context) error {
	const query = `CREATE TABLE IF NOT EXISTS publication_attachment (
		release_id CHAR(36) NOT NULL,
		file_id CHAR(36) NOT NULL,
		copy_file_id CHAR(36) NOT NULL,
		file_name VARCHAR(255) NOT NULL,
		content_type VARCHAR(255) NOT NULL,
		size_bytes BIGINT NOT NULL,
		PRIMARY KEY (release_id, file_id),
		CONSTRAINT publication_attachment_release_fk FOREIGN KEY (release_id) REFERENCES publication_release(id)
	) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	_, err := v.client.ExecContext(ctx, "create-publication-attachment", query)
	return err
}
