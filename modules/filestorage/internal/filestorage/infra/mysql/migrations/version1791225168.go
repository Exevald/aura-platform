package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225168 struct{ db mysql.ClientContext }

func Version1791225168(db mysql.ClientContext) migrator.Migration { return &version1791225168{db} }

func (*version1791225168) Version() int64 {
	return 1791225168
}

func (*version1791225168) Description() string {
	return "Create file metadata"
}

func (v *version1791225168) Up(ctx context.Context) error {
	_, e := v.db.ExecContext(ctx, "filestorage.schema", `CREATE TABLE IF NOT EXISTS stored_file(id BINARY(16) NOT NULL PRIMARY KEY,workspace_id BINARY(16) NOT NULL,owner_user_id BINARY(16) NOT NULL,object_key VARCHAR(512) NOT NULL UNIQUE,file_name VARCHAR(255) NOT NULL,content_type VARCHAR(255) NOT NULL,size_bytes BIGINT NOT NULL,checksum_sha256 CHAR(64) NOT NULL,created_at DATETIME(6) NOT NULL)`)
	return e
}
