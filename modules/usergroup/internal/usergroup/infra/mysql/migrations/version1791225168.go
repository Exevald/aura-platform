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
	return "Create user groups and members"
}

func (m *version1791225168) Up(ctx context.Context) error {
	const (
		sqlCreateUserGroupQuery = `
			CREATE TABLE IF NOT EXISTS user_group (
			    id BINARY(16) NOT NULL PRIMARY KEY,
			    owner_user_id BINARY(16) NOT NULL,
			    workspace_id BINARY(16) NOT NULL,
			    name VARCHAR(120) NOT NULL,
			    created_at DATETIME(6) NOT NULL,
			    updated_at DATETIME(6) NOT NULL,
			    INDEX user_group_owner_idx(owner_user_id),
			    INDEX user_group_workspace_idx(workspace_id)
			) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
		`
		sqlCreateUserGroupMemberQuery = `
			CREATE TABLE IF NOT EXISTS user_group_member (
			    group_id BINARY(16) NOT NULL,
			    user_id BINARY(16) NOT NULL,
			    created_at DATETIME(6) NOT NULL,
			    PRIMARY KEY(group_id, user_id),
			    INDEX user_group_member_user_idx(user_id),
			    CONSTRAINT user_group_member_group_fk FOREIGN KEY(group_id) REFERENCES user_group(id) ON DELETE CASCADE
			) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci
		`
	)

	if _, err := m.client.ExecContext(ctx, "create-user-group", sqlCreateUserGroupQuery); err != nil {
		return err
	}
	_, err := m.client.ExecContext(ctx, "create-user-group-member", sqlCreateUserGroupMemberQuery)
	return err
}
