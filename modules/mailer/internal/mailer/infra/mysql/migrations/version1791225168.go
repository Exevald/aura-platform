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
	return "Create mail jobs"
}

func (v *version1791225168) Up(ctx context.Context) error {
	_, e := v.db.ExecContext(ctx, "mailer.schema", `CREATE TABLE IF NOT EXISTS mail_job(id BINARY(16) NOT NULL PRIMARY KEY,idempotency_key VARCHAR(191) NOT NULL UNIQUE,recipient VARCHAR(320) NOT NULL,subject VARCHAR(255) NOT NULL,body TEXT NOT NULL,status ENUM('pending','sent','failed') NOT NULL,attempts INT NOT NULL DEFAULT 0,last_error TEXT NULL,created_at DATETIME(6) NOT NULL,updated_at DATETIME(6) NOT NULL)`)
	return e
}
