package repository

import (
	"context"
	"database/sql"
	"errors"
	"planner/internal/planner/domain"
	"time"
	"uuid"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type repo struct {
	ctx context.Context
	db  sdkmysql.ClientContext
}

func New(ctx context.Context, db sdkmysql.ClientContext) domain.Repository { return &repo{ctx, db} }

type projectRow struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type taskRow struct {
	ID              uuid.UUID      `db:"id"`
	ProjectID       uuid.UUID      `db:"project_id"`
	Title           string         `db:"title"`
	Description     string         `db:"description"`
	Status          string         `db:"status"`
	Priority        string         `db:"priority"`
	AssigneeUserID  sql.NullString `db:"assignee_user_id"`
	DueAt           sql.NullTime   `db:"due_at"`
	CalendarID      sql.NullString `db:"calendar_id"`
	CalendarEventID sql.NullString `db:"calendar_event_id"`
	Timezone        string         `db:"timezone"`
	SyncToCalendar  bool           `db:"sync_to_calendar"`
	Version         int64          `db:"version"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

func (r *repo) StoreProject(v domain.Project) error {
	_, e := r.db.ExecContext(r.ctx, "project.store", `INSERT INTO project(id,workspace_id,name,created_at,updated_at) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE name=VALUES(name),updated_at=VALUES(updated_at)`, v.ID, v.WorkspaceID, v.Name, v.CreatedAt, v.UpdatedAt)
	return e
}

func (r *repo) FindProject(id uuid.UUID) (domain.Project, error) {
	var v projectRow
	e := r.db.GetContext(r.ctx, "project.find", &v, `SELECT id,workspace_id,name,created_at,updated_at FROM project WHERE id=?`, id)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return domain.Project(v), e
}

func (r *repo) ListProjects(w uuid.UUID) ([]domain.Project, error) {
	var rows []projectRow
	e := r.db.SelectContext(r.ctx, "project.list", &rows, `SELECT id,workspace_id,name,created_at,updated_at FROM project WHERE workspace_id=? ORDER BY name`, w)
	out := make([]domain.Project, 0, len(rows))
	for _, v := range rows {
		out = append(out, domain.Project(v))
	}
	return out, e
}

func (r *repo) DeleteProject(w, id uuid.UUID) error {
	res, e := r.db.ExecContext(r.ctx, "project.delete", `DELETE FROM project WHERE id=? AND workspace_id=?`, id, w)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

const taskColumns = `id,project_id,title,description,status,priority,assignee_user_id,due_at,calendar_id,calendar_event_id,timezone,sync_to_calendar,version,created_at,updated_at`

func (r *repo) StoreTask(v domain.Task, expected int64) (domain.Task, error) {
	args := func(x uuid.UUID) any {
		if x == uuid.Nil() {
			return nil
		}
		return x
	}
	var due any
	if v.DueAt != nil {
		due = *v.DueAt
	}
	if expected == 0 {
		_, e := r.db.ExecContext(r.ctx, "task.insert", `INSERT INTO task(id,project_id,title,description,status,priority,assignee_user_id,due_at,calendar_id,calendar_event_id,timezone,sync_to_calendar,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.ProjectID, v.Title, v.Description, v.Status, v.Priority, args(v.AssigneeUserID), due, args(v.CalendarID), args(v.CalendarEventID), v.Timezone, v.SyncToCalendar, v.Version, v.CreatedAt, v.UpdatedAt)
		return v, e
	}
	res, e := r.db.ExecContext(r.ctx, "task.update", `UPDATE task SET title=?,description=?,status=?,priority=?,assignee_user_id=?,due_at=?,calendar_id=?,calendar_event_id=?,timezone=?,sync_to_calendar=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Title, v.Description, v.Status, v.Priority, args(v.AssigneeUserID), due, args(v.CalendarID), args(v.CalendarEventID), v.Timezone, v.SyncToCalendar, v.UpdatedAt, v.ID, expected)
	if e != nil {
		return domain.Task{}, e
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.Task{}, domain.ErrConflict
	}
	v.Version = expected + 1
	return v, nil
}

func (r *repo) FindTask(id uuid.UUID) (domain.Task, error) {
	var v taskRow
	e := r.db.GetContext(r.ctx, "task.find", &v, `SELECT `+taskColumns+` FROM task WHERE id=?`, id)
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrNotFound
	}
	if e != nil {
		return domain.Task{}, e
	}
	return v.domain(), nil
}

func (r *repo) ListTasks(p uuid.UUID) ([]domain.Task, error) {
	var rows []taskRow
	e := r.db.SelectContext(r.ctx, "task.list", &rows, `SELECT `+taskColumns+` FROM task WHERE project_id=? ORDER BY status,due_at,created_at`, p)
	out := make([]domain.Task, 0, len(rows))
	for _, v := range rows {
		out = append(out, v.domain())
	}
	return out, e
}

func (r *repo) DeleteTask(id uuid.UUID) error {
	res, e := r.db.ExecContext(r.ctx, "task.delete", `DELETE FROM task WHERE id=?`, id)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

func (v taskRow) domain() domain.Task {
	out := domain.Task{ID: v.ID, ProjectID: v.ProjectID, Title: v.Title, Description: v.Description, Status: v.Status, Priority: v.Priority, Timezone: v.Timezone, SyncToCalendar: v.SyncToCalendar, Version: v.Version, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	parse := func(x sql.NullString) uuid.UUID {
		if !x.Valid {
			return uuid.Nil()
		}
		id, _ := uuid.Parse(x.String)
		return id
	}
	out.AssigneeUserID = parse(v.AssigneeUserID)
	out.CalendarID = parse(v.CalendarID)
	out.CalendarEventID = parse(v.CalendarEventID)
	if v.DueAt.Valid {
		t := v.DueAt.Time
		out.DueAt = &t
	}
	return out
}
