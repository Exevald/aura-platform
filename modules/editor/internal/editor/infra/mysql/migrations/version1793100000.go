package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type attachmentMigration struct{ client mysql.ClientContext }

func Version1793100000(c mysql.ClientContext) migrator.Migration { return &attachmentMigration{c} }

func (*attachmentMigration) Version() int64 {
	return 1793100000
}

func (*attachmentMigration) Description() string {
	return "Create editor document attachments"
}

func (m *attachmentMigration) Up(ctx context.Context) error {
	const sqlCreateDocumentAttachmentQuery = `
		CREATE TABLE IF NOT EXISTS document_attachment(
		    document_id BINARY(16) NOT NULL,
		    file_id CHAR(36) NOT NULL,
		    workspace_id BINARY(16) NOT NULL,
		    file_name VARCHAR(255) NOT NULL,
		    content_type VARCHAR(255) NOT NULL,
		    size_bytes BIGINT NOT NULL,
		    created_at DATETIME(6) NOT NULL,
		    PRIMARY KEY(document_id, file_id),
		    CONSTRAINT document_attachment_document_fk FOREIGN KEY(document_id) REFERENCES document(id) ON DELETE CASCADE
		) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
	`

	_, e := m.client.ExecContext(ctx, "create-document-attachment", sqlCreateDocumentAttachmentQuery)
	return e
}
