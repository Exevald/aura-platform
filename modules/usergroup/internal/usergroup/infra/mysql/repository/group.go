package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"usergroup/internal/usergroup/domain"
)

func NewGroup(ctx context.Context, client mysql.ClientContext) domain.GroupRepository {
	return &groupRepository{ctx: ctx, client: client}
}

type groupRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func (*groupRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *groupRepository) Store(group domain.Group) error {
	const query = `
		INSERT INTO user_group(
		    id,
		    owner_user_id,
		    workspace_id,
		    name,
		    created_at,
		    updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    name = VALUES (name),
		    updated_at = VALUES (updated_at)
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"group.store",
		query,
		group.ID,
		group.OwnerUserID,
		group.WorkspaceID,
		group.Name,
		group.CreatedAt,
		group.UpdatedAt,
	)
	return err
}

func (r *groupRepository) Find(id uuid.UUID) (domain.Group, error) {
	const query = `
		SELECT
		    id,
		    owner_user_id,
		    workspace_id,
		    name,
		    created_at,
		    updated_at
		FROM user_group
		WHERE id = ?
	`
	var row groupRow
	err := r.client.GetContext(r.ctx, "group.find", &row, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Group{}, domain.ErrGroupNotFound
	}
	if err != nil {
		return domain.Group{}, err
	}
	return row.domain(), nil
}

func (r *groupRepository) ListByOwner(owner, workspace uuid.UUID) ([]domain.Group, error) {
	const query = `
		SELECT
		    id,
		    owner_user_id,
		    workspace_id,
		    name,
		    created_at,
		    updated_at
		FROM user_group
		WHERE owner_user_id = ? AND workspace_id = ?
		ORDER BY name
	`
	var rows []groupRow
	if err := r.client.SelectContext(r.ctx, "group.list", &rows, query, owner, workspace); err != nil {
		return nil, err
	}
	result := make([]domain.Group, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.domain())
	}
	return result, nil
}

func (r *groupRepository) Delete(id uuid.UUID) error {
	const sqlGroupDeleteQuery = `
		DELETE FROM user_group
		WHERE id = ?
	`

	res, err := r.client.ExecContext(r.ctx, "group.delete", sqlGroupDeleteQuery, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrGroupNotFound
	}
	return nil
}

type groupRow struct {
	ID          uuid.UUID `db:"id"`
	OwnerUserID uuid.UUID `db:"owner_user_id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (r groupRow) domain() domain.Group {
	return domain.Group{
		ID:          r.ID,
		OwnerUserID: r.OwnerUserID,
		WorkspaceID: r.WorkspaceID,
		Name:        r.Name,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
