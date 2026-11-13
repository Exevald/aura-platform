package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225168 struct {
	db mysql.ClientContext
}

func Version1791225168(db mysql.ClientContext) migrator.Migration {
	return &version1791225168{db}
}

func (*version1791225168) Version() int64 {
	return 1791225168
}

func (*version1791225168) Description() string {
	return "Create projects and tasks"
}

func (v *version1791225168) Up(ctx context.Context) error {
	schema := `
		CREATE TABLE IF NOT EXISTS project(
		    id BINARY(16) NOT NULL PRIMARY KEY,
		    workspace_id BINARY(16) NOT NULL,
		    name VARCHAR(120) NOT NULL,
		    created_at DATETIME(6) NOT NULL,
		    updated_at DATETIME(6) NOT NULL,
		INDEX project_workspace_idx(workspace_id));
		CREATE TABLE IF NOT EXISTS task(
		    id BINARY(16) NOT NULL PRIMARY KEY,
		    project_id BINARY(16) NOT NULL,
		    title VARCHAR(255) NOT NULL,
		    description TEXT NOT NULL,
		    status ENUM('todo','in_progress','done') NOT NULL,
		    priority ENUM('low','normal','high') NOT NULL,
		    assignee_user_id BINARY(16) NULL,
		    due_at DATETIME(6) NULL,
		    calendar_id BINARY(16) NULL,
		    calendar_event_id BINARY(16) NULL,
		    timezone VARCHAR(80) NOT NULL DEFAULT 'UTC',
		    sync_to_calendar BOOLEAN NOT NULL DEFAULT FALSE,
		    version BIGINT NOT NULL,created_at DATETIME(6) NOT NULL,
		    updated_at DATETIME(6) NOT NULL,
		INDEX task_project_status_idx(project_id,status),
		CONSTRAINT task_project_fk FOREIGN KEY(project_id) REFERENCES project(id) ON DELETE CASCADE)`
	for i, statement := range strings.Split(schema, ";") {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := v.db.ExecContext(ctx, fmt.Sprintf("planner.schema.%d", i), statement); err != nil {
			return err
		}
	}
	return nil
}
