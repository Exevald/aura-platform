package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

func Version1791292526(client mysql.ClientContext) migrator.Migration {
	return &uniqueEmailMigration{client: client}
}

type uniqueEmailMigration struct{ client mysql.ClientContext }

func (*uniqueEmailMigration) Version() int64 { return 1791292526 }
func (*uniqueEmailMigration) Description() string {
	return "Ensure user email is unique"
}

func (m *uniqueEmailMigration) Up(ctx context.Context) error {
	var count int
	if err := m.client.GetContext(ctx, "find-user-email-index", &count, `
SELECT COUNT(*)
FROM information_schema.statistics
WHERE table_schema = DATABASE() AND table_name = 'user' AND index_name = 'user_email_uq'
`); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := m.client.ExecContext(ctx, "create-user-email-index", `ALTER TABLE user ADD UNIQUE KEY user_email_uq (email)`)
	return err
}
