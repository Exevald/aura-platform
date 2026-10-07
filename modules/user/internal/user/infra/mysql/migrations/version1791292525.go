package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1791292525(client mysql.ClientContext) migrator.Migration {
	return &version1791292525{
		client: client,
	}
}

type version1791292525 struct {
	client mysql.ClientContext
}

func (v *version1791292525) Version() int64 {
	return 1791292525
}

func (v *version1791292525) Description() string {
	return "Create user table"
}

func (v *version1791292525) Up(ctx context.Context) error {
	_, err := v.client.ExecContext(ctx, "create-user-table", `
		CREATE TABLE IF NOT EXISTS user (
			id BINARY(16) NOT NULL PRIMARY KEY,
		    login VARCHAR(255) NOT NULL,
		    first_name VARCHAR(255) NOT NULL,
		    last_name VARCHAR(255) NOT NULL,
		    email VARCHAR(255) NOT NULL,
		    deleted_at DATETIME NULL DEFAULT NULL
		) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci
`)
	return err
}
