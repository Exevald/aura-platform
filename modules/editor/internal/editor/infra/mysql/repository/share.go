package repository

import (
	"context"
	"database/sql"
	"editor/internal/editor/domain/model"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type shareRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type shareRow struct {
	ID            uuid.UUID         `db:"id"`
	ModuleID      uuid.UUID         `db:"module_id"`
	GroupID       uuid.UUID         `db:"group_id"`
	OwnerUserID   uuid.UUID         `db:"owner_user_id"`
	Status        model.ShareStatus `db:"status"`
	Attempt       int               `db:"attempt"`
	FailureReason sql.NullString    `db:"failure_reason"`
	CreatedAt     time.Time         `db:"created_at"`
	UpdatedAt     time.Time         `db:"updated_at"`
}

func NewShare(ctx context.Context, client mysql.ClientContext) model.ModuleShareRepository {
	return &shareRepository{ctx: ctx, client: client}
}

func (*shareRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *shareRepository) Store(v model.ModuleShare) error {
	const sqlModuleShareStoreQuery = `
		INSERT INTO module_share(
		    id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    attempt,
		    failure_reason,
		    created_at,
		    updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    status = VALUES (status),
		    attempt = VALUES (attempt),
		    failure_reason = VALUES (failure_reason),
		    updated_at = VALUES (updated_at)
	`

	_, err := r.client.ExecContext(r.ctx, "module-share.store", sqlModuleShareStoreQuery, v.ID, v.ModuleID, v.GroupID, v.OwnerUserID, v.Status, v.Attempt, nullable(v.FailureReason), v.CreatedAt, v.UpdatedAt)
	return err
}

func (r *shareRepository) Find(id uuid.UUID) (model.ModuleShare, error) {
	const sqlModuleShareFindQuery = `
		SELECT
		    id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    attempt,
		    failure_reason,
		    created_at,
		    updated_at
		FROM module_share
		WHERE id = ?
	`

	var row shareRow
	err := r.client.GetContext(r.ctx, "module-share.find", &row, sqlModuleShareFindQuery, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ModuleShare{}, model.ErrShareNotFound
	}
	return row.domain(), err
}

func (r *shareRepository) FindByOwnerGroup(ownerID, groupID uuid.UUID) (model.ModuleShare, error) {
	const sqlModuleShareFindOwnerGroupQuery = `
		SELECT
		    id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    attempt,
		    failure_reason,
		    created_at,
		    updated_at
		FROM module_share
		WHERE owner_user_id = ? AND group_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var row shareRow
	err := r.client.GetContext(r.ctx, "module-share.find-owner-group", &row, sqlModuleShareFindOwnerGroupQuery, ownerID, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ModuleShare{}, model.ErrShareNotFound
	}
	return row.domain(), err
}

func (r *shareRepository) ListByOwner(ownerID uuid.UUID) ([]model.ModuleShare, error) {
	const sqlModuleShareListQuery = `
		SELECT
		    id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    attempt,
		    failure_reason,
		    created_at,
		    updated_at
		FROM module_share
		WHERE owner_user_id = ?
		ORDER BY created_at DESC
	`

	var rows []shareRow
	if err := r.client.SelectContext(r.ctx, "module-share.list", &rows, sqlModuleShareListQuery, ownerID); err != nil {
		return nil, err
	}
	return convertShares(rows), nil
}

func (r *shareRepository) ListPendingBefore(deadline time.Time) ([]model.ModuleShare, error) {
	const sqlModuleSharePendingQuery = `
		SELECT
		    id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    attempt,
		    failure_reason,
		    created_at,
		    updated_at
		FROM module_share
		WHERE status = 'pending' AND updated_at <= ?
		FOR UPDATE SKIP LOCKED
	`

	var rows []shareRow
	if err := r.client.SelectContext(r.ctx, "module-share.pending", &rows, sqlModuleSharePendingQuery, deadline); err != nil {
		return nil, err
	}
	return convertShares(rows), nil
}

func (r *shareRepository) ListCompensatingBefore(deadline time.Time) ([]model.ModuleShare, error) {
	const sqlModuleShareCompensatingQuery = `
		SELECT
		    id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    attempt,
		    failure_reason,
		    created_at,
		    updated_at
		FROM module_share
		WHERE status = 'compensating' AND updated_at <= ?
		FOR UPDATE SKIP LOCKED
	`

	var rows []shareRow
	if err := r.client.SelectContext(r.ctx, "module-share.compensating", &rows, sqlModuleShareCompensatingQuery, deadline); err != nil {
		return nil, err
	}
	return convertShares(rows), nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (r shareRow) domain() model.ModuleShare {
	return model.ModuleShare{ID: r.ID, ModuleID: r.ModuleID, GroupID: r.GroupID, OwnerUserID: r.OwnerUserID, Status: r.Status, Attempt: r.Attempt, FailureReason: r.FailureReason.String, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func convertShares(rows []shareRow) []model.ModuleShare {
	result := make([]model.ModuleShare, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.domain())
	}
	return result
}
