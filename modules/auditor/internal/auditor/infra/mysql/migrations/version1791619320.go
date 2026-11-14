package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791619320 struct {
	client mysql.ClientContext
}

func Version1791619320(client mysql.ClientContext) migrator.Migration {
	return &version1791619320{
		client: client,
	}
}

func (v *version1791619320) Version() int64 {
	return 1791619320
}

func (v *version1791619320) Description() string {
	return "Create audit records table"
}

func (v *version1791619320) Up(ctx context.Context) error {
	const sqlCreateAuditRecordTableQuery = `
		CREATE TABLE IF NOT EXISTS audit_record (
		    id BINARY(16) NOT NULL PRIMARY KEY,
		    event_id VARCHAR(255) NOT NULL,
		    event_type VARCHAR(255) NOT NULL,
		    producer VARCHAR(255) NOT NULL,
		    correlation_id VARCHAR(255) NOT NULL,
		    actor_id VARCHAR(255) NOT NULL,
		    resource_type VARCHAR(255) NOT NULL,
		    resource_id VARCHAR(255) NOT NULL,
		    payload LONGBLOB NOT NULL,
		    occurred_at DATETIME(6) NOT NULL,
		    received_at DATETIME(6) NOT NULL,
		    UNIQUE KEY audit_record_event_uidx(event_id)
		) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
	`

	_, err := v.client.ExecContext(ctx, "create-audit-record-table", sqlCreateAuditRecordTableQuery)
	return err
}
