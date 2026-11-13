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

type moduleShareOperationRow struct {
	ShareID     binaryUUID              `db:"share_id"`
	ModuleID    binaryUUID              `db:"module_id"`
	GroupID     binaryUUID              `db:"group_id"`
	OwnerUserID binaryUUID              `db:"owner_user_id"`
	Status      model.ModuleShareStatus `db:"status"`
	CreatedAt   time.Time               `db:"created_at"`
	UpdatedAt   time.Time               `db:"updated_at"`
}

func NewModuleShareOperation(ctx context.Context, client mysql.ClientContext) model.ModuleShareRepository {
	return &moduleShareOperationRepository{ctx: ctx, client: client}
}

func (r *moduleShareOperationRepository) ListByGroup(group uuid.UUID) ([]model.ModuleShare, error) {
	var rows []moduleShareOperationRow
	if err := r.client.SelectContext(r.ctx, "module-share-operation.list-group", &rows, selectOperation+` WHERE group_id=? AND status='granted'`, newBinaryUUID(group)); err != nil {
		return nil, err
	}
	result := make([]model.ModuleShare, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.domain())
	}
	return result, nil
}

func (r *moduleShareOperationRepository) ListByOwner(owner uuid.UUID) ([]model.ModuleShare, error) {
	var rows []moduleShareOperationRow
	if err := r.client.SelectContext(r.ctx, "module-share-operation.list-owner", &rows, selectOperation+` WHERE owner_user_id=? AND status='granted'`, newBinaryUUID(owner)); err != nil {
		return nil, err
	}
	result := make([]model.ModuleShare, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.domain())
	}
	return result, nil
}

func (r *moduleShareOperationRepository) Find(id uuid.UUID) (model.ModuleShare, error) {
	var row moduleShareOperationRow
	err := r.client.GetContext(r.ctx, "module-share-operation.find", &row, selectOperation+` WHERE share_id=?`, newBinaryUUID(id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.ModuleShare{}, model.ErrModuleShareOperationNotFound
	}
	return row.domain(), err
}

func (r *moduleShareOperationRepository) Store(op model.ModuleShare) error {
	_, err := r.client.ExecContext(r.ctx, "module-share-operation.store", `INSERT INTO module_share_operation(share_id,module_id,group_id,owner_user_id,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE status=VALUES(status),updated_at=VALUES(updated_at)`, newBinaryUUID(op.ShareID), newBinaryUUID(op.ModuleID), newBinaryUUID(op.GroupID), newBinaryUUID(op.OwnerUserID), op.Status, op.CreatedAt, op.UpdatedAt)
	return err
}

const selectOperation = `SELECT share_id,module_id,group_id,owner_user_id,status,created_at,updated_at FROM module_share_operation`

func (r moduleShareOperationRow) domain() model.ModuleShare {
	return model.ModuleShare{ShareID: r.ShareID.UUID(), ModuleID: r.ModuleID.UUID(), GroupID: r.GroupID.UUID(), OwnerUserID: r.OwnerUserID.UUID(), Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
