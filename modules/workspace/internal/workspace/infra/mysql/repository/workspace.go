package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"uuid"
	"workspace/internal/workspace/domain"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type repo struct {
	ctx context.Context
	db  mysql.ClientContext
}

func New(ctx context.Context, db mysql.ClientContext) domain.Repository {
	return &repo{ctx: ctx, db: db}
}

type workspaceRow struct {
	ID        uuid.UUID   `db:"id"`
	Kind      domain.Kind `db:"kind"`
	Name      string      `db:"name"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

type memberRow struct {
	WorkspaceID uuid.UUID   `db:"workspace_id"`
	UserID      uuid.UUID   `db:"user_id"`
	Role        domain.Role `db:"role"`
	CreatedAt   time.Time   `db:"created_at"`
}

func (r *repo) Store(v domain.Workspace) error {
	_, e := r.db.ExecContext(r.ctx, "workspace.store", `INSERT INTO workspace(id,kind,name,created_at,updated_at) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE name=VALUES(name),updated_at=VALUES(updated_at)`, v.ID, v.Kind, v.Name, v.CreatedAt, v.UpdatedAt)
	return e
}

func (r *repo) Find(id uuid.UUID) (domain.Workspace, error) {
	var v workspaceRow
	e := r.db.GetContext(r.ctx, "workspace.find", &v, `SELECT id,kind,name,created_at,updated_at FROM workspace WHERE id=?`, id)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return domain.Workspace(v), e
}

func (r *repo) ListByUser(user uuid.UUID) ([]domain.Workspace, error) {
	var rows []workspaceRow
	e := r.db.SelectContext(r.ctx, "workspace.list", &rows, `SELECT w.id,w.kind,w.name,w.created_at,w.updated_at FROM workspace w JOIN workspace_member m ON m.workspace_id=w.id WHERE m.user_id=? ORDER BY w.kind,w.name`, user)
	out := make([]domain.Workspace, 0, len(rows))
	for _, v := range rows {
		out = append(out, domain.Workspace(v))
	}
	return out, e
}

func (r *repo) Delete(id uuid.UUID) error {
	res, e := r.db.ExecContext(r.ctx, "workspace.delete", `DELETE FROM workspace WHERE id=?`, id)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

func (r *repo) StoreMember(v domain.Member) error {
	_, e := r.db.ExecContext(r.ctx, "workspace.member.store", `INSERT INTO workspace_member(workspace_id,user_id,role,created_at) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE role=VALUES(role)`, v.WorkspaceID, v.UserID, v.Role, v.CreatedAt)
	return e
}

func (r *repo) DeleteMember(w, u uuid.UUID) error {
	res, e := r.db.ExecContext(r.ctx, "workspace.member.delete", `DELETE FROM workspace_member WHERE workspace_id=? AND user_id=?`, w, u)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

func (r *repo) FindMember(w, u uuid.UUID) (domain.Member, error) {
	var v memberRow
	e := r.db.GetContext(r.ctx, "workspace.member.find", &v, `SELECT workspace_id,user_id,role,created_at FROM workspace_member WHERE workspace_id=? AND user_id=?`, w, u)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return domain.Member(v), e
}

func (r *repo) ListMembers(w uuid.UUID) ([]domain.Member, error) {
	var rows []memberRow
	e := r.db.SelectContext(r.ctx, "workspace.member.list", &rows, `SELECT workspace_id,user_id,role,created_at FROM workspace_member WHERE workspace_id=? ORDER BY created_at`, w)
	out := make([]domain.Member, 0, len(rows))
	for _, v := range rows {
		out = append(out, domain.Member(v))
	}
	return out, e
}

func (r *repo) StoreInvitation(v domain.Invitation) error {
	var user any
	if v.UserID != uuid.Nil() {
		user = v.UserID
	}
	_, e := r.db.ExecContext(r.ctx, "workspace.invitation.store", `INSERT INTO workspace_invitation(id,workspace_id,user_id,email,role,token,expires_at,created_at) VALUES(?,?,?,?,?,?,?,?)`, v.ID, v.WorkspaceID, user, v.Email, v.Role, v.Token, v.ExpiresAt, v.CreatedAt)
	return e
}

type invitationRow struct {
	ID          uuid.UUID   `db:"id"`
	WorkspaceID uuid.UUID   `db:"workspace_id"`
	UserID      *uuid.UUID  `db:"user_id"`
	Email       string      `db:"email"`
	Role        domain.Role `db:"role"`
	Token       string      `db:"token"`
	ExpiresAt   time.Time   `db:"expires_at"`
	CreatedAt   time.Time   `db:"created_at"`
}

func invitationFromRow(v invitationRow) domain.Invitation {
	out := domain.Invitation{ID: v.ID, WorkspaceID: v.WorkspaceID, Email: v.Email, Role: v.Role, Token: v.Token, ExpiresAt: v.ExpiresAt, CreatedAt: v.CreatedAt}
	if v.UserID != nil {
		out.UserID = *v.UserID
	}
	return out
}

func (r *repo) FindInvitationByToken(token string) (domain.Invitation, error) {
	var v invitationRow
	e := r.db.GetContext(r.ctx, "workspace.invitation.find", &v, `SELECT id,workspace_id,user_id,email,role,token,expires_at,created_at FROM workspace_invitation WHERE token=?`, token)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return invitationFromRow(v), e
}

func (r *repo) ListInvitations(workspace uuid.UUID) ([]domain.Invitation, error) {
	var rows []invitationRow
	e := r.db.SelectContext(r.ctx, "workspace.invitation.list", &rows, `SELECT id,workspace_id,user_id,email,role,token,expires_at,created_at FROM workspace_invitation WHERE workspace_id=? ORDER BY created_at DESC`, workspace)
	out := make([]domain.Invitation, 0, len(rows))
	for _, v := range rows {
		out = append(out, invitationFromRow(v))
	}
	return out, e
}

func (r *repo) DeleteInvitation(workspace, id uuid.UUID) error {
	res, e := r.db.ExecContext(r.ctx, "workspace.invitation.delete", `DELETE FROM workspace_invitation WHERE id=? AND workspace_id=?`, id, workspace)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}
