package repository

import (
	"access/internal/access/domain/model"
	"context"
	"database/sql"
	"errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type moduleShareOperationRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func NewModuleShareOperation(ctx context.Context, client mysql.ClientContext) model.ModuleShareRepository {
	return &moduleShareOperationRepository{ctx: ctx, client: client}
}

func (r *moduleShareOperationRepository) ListByGroup(group uuid.UUID) ([]model.ModuleShare, error) {
	const sqlModuleShareOperationListGroupQuery = `
		SELECT
		    share_id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    created_at,
		    updated_at
		FROM module_share_operation
		WHERE group_id = ? AND status = 'granted'
	`

	var rows []moduleShareOperationRow
	if err := r.client.SelectContext(r.ctx, "module-share-operation.list-group", &rows, sqlModuleShareOperationListGroupQuery, newBinaryUUID(group)); err != nil {
		return nil, err
	}
	result := make([]model.ModuleShare, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.toDomainModel())
	}
	return result, nil
}

func (r *moduleShareOperationRepository) ListByOwner(owner uuid.UUID) ([]model.ModuleShare, error) {
	const sqlModuleShareOperationListOwnerQuery = `
		SELECT
		    share_id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    created_at,
		    updated_at
		FROM module_share_operation
		WHERE owner_user_id = ? AND status = 'granted'
	`

	var rows []moduleShareOperationRow
	if err := r.client.SelectContext(r.ctx, "module-share-operation.list-owner", &rows, sqlModuleShareOperationListOwnerQuery, newBinaryUUID(owner)); err != nil {
		return nil, err
	}
	result := make([]model.ModuleShare, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.toDomainModel())
	}
	return result, nil
}

func (r *moduleShareOperationRepository) Find(id uuid.UUID) (model.ModuleShare, error) {
	const sqlModuleShareOperationFindQuery = `
		SELECT
		    share_id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    created_at,
		    updated_at
		FROM module_share_operation
		WHERE share_id = ?
	`

	var row moduleShareOperationRow
	err := r.client.GetContext(r.ctx, "module-share-operation.find", &row, sqlModuleShareOperationFindQuery, newBinaryUUID(id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.ModuleShare{}, model.ErrModuleShareOperationNotFound
	}
	return row.toDomainModel(), err
}

func (r *moduleShareOperationRepository) Store(op model.ModuleShare) error {
	const sqlModuleShareOperationStoreQuery = `
		INSERT INTO module_share_operation(
		    share_id,
		    module_id,
		    group_id,
		    owner_user_id,
		    status,
		    created_at,
		    updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    status = VALUES (status),
		    updated_at = VALUES (updated_at)
	`

	_, err := r.client.ExecContext(r.ctx, "module-share-operation.store", sqlModuleShareOperationStoreQuery, newBinaryUUID(op.ShareID), newBinaryUUID(op.ModuleID), newBinaryUUID(op.GroupID), newBinaryUUID(op.OwnerUserID), op.Status, op.CreatedAt, op.UpdatedAt)
	return err
}

type moduleShareOperationRow struct {
	ShareID     binaryUUID              `db:"share_id"`
	ModuleID    binaryUUID              `db:"module_id"`
	GroupID     binaryUUID              `db:"group_id"`
	OwnerUserID binaryUUID              `db:"owner_user_id"`
	Status      model.ModuleShareStatus `db:"status"`
	CreatedAt   time.Time               `db:"created_at"`
	UpdatedAt   time.Time               `db:"updated_at"`
}

func (r *moduleShareOperationRow) toDomainModel() model.ModuleShare {
	return model.ModuleShare{
		ShareID:     r.ShareID.UUID(),
		ModuleID:    r.ModuleID.UUID(),
		GroupID:     r.GroupID.UUID(),
		OwnerUserID: r.OwnerUserID.UUID(),
		Status:      r.Status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
