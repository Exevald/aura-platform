package migrations

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225168 struct{ client mysql.ClientContext }

func Version1791225168(c mysql.ClientContext) migrator.Migration { return &version1791225168{c} }

func (*version1791225168) Version() int64 {
	return 1791225168
}

func (*version1791225168) Description() string {
	return "Create workspaces and memberships"
}

func (v *version1791225168) Up(ctx context.Context) error {
	statements := []struct{ name, sql string }{
		{"create-workspace", `CREATE TABLE IF NOT EXISTS workspace(id BINARY(16) NOT NULL PRIMARY KEY,kind ENUM('personal','team') NOT NULL,name VARCHAR(120) NOT NULL,created_at DATETIME(6) NOT NULL,updated_at DATETIME(6) NOT NULL) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},
		{"create-workspace-member", `CREATE TABLE IF NOT EXISTS workspace_member(workspace_id BINARY(16) NOT NULL,user_id BINARY(16) NOT NULL,role ENUM('owner','editor','viewer') NOT NULL,created_at DATETIME(6) NOT NULL,PRIMARY KEY(workspace_id,user_id),INDEX workspace_member_user_idx(user_id),CONSTRAINT workspace_member_workspace_fk FOREIGN KEY(workspace_id) REFERENCES workspace(id) ON DELETE CASCADE) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},
		{"create-workspace-invitation", `CREATE TABLE IF NOT EXISTS workspace_invitation(id BINARY(16) NOT NULL PRIMARY KEY,workspace_id BINARY(16) NOT NULL,user_id BINARY(16) NULL,email VARCHAR(320) NOT NULL,role ENUM('editor','viewer') NOT NULL,token CHAR(64) NOT NULL,expires_at DATETIME(6) NOT NULL,created_at DATETIME(6) NOT NULL,UNIQUE KEY workspace_invitation_token_uidx(token),INDEX workspace_invitation_workspace_idx(workspace_id),CONSTRAINT workspace_invitation_workspace_fk FOREIGN KEY(workspace_id) REFERENCES workspace(id) ON DELETE CASCADE) ENGINE=InnoDB CHARACTER SET=utf8mb4 COLLATE=utf8mb4_unicode_ci`},
	}
	for _, statement := range statements {
		if _, err := v.client.ExecContext(ctx, statement.name, statement.sql); err != nil {
			return err
		}
	}
	return nil
}
