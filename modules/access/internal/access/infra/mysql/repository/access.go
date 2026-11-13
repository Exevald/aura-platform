package repository

import (
	"context"
	"database/sql"
	stderrors "errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"

	"access/internal/access/domain/model"
)

type moduleAccessRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type moduleAccessRow struct {
	ID        binaryUUID   `db:"id"`
	UserID    binaryUUID   `db:"user_id"`
	ModuleID  binaryUUID   `db:"module_id"`
	Effect    model.Effect `db:"effect"`
	SourceID  string       `db:"source_id"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

func NewModuleAccessRepository(
	ctx context.Context,
	client mysql.ClientContext,
) model.ModuleAccessRepository {
	return &moduleAccessRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *moduleAccessRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *moduleAccessRepository) Store(access model.ModuleAccess) error {
	const query = `
		INSERT INTO module_access
			(id, user_id, module_id, effect, source_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			module_id = VALUES(module_id),
			effect = VALUES(effect),
			source_id = VALUES(source_id),
			updated_at = VALUES(updated_at)
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"store-module-access",
		query,
		newBinaryUUID(access.ID),
		newBinaryUUID(access.UserID),
		newBinaryUUID(access.ModuleID),
		access.Effect,
		access.SourceID,
		access.CreatedAt,
		access.UpdatedAt,
	)
	return errors.WithStack(err)
}

func (r *moduleAccessRepository) Find(id uuid.UUID) (model.ModuleAccess, error) {
	const query = `
		SELECT id, user_id, module_id, effect, source_id, created_at, updated_at
		FROM module_access
		WHERE id = ?
	`
	var row moduleAccessRow
	err := r.client.GetContext(r.ctx, "find-module-access", &row, query, newBinaryUUID(id))
	if stderrors.Is(err, sql.ErrNoRows) {
		return model.ModuleAccess{}, errors.WithStack(model.ErrModuleAccessNotFound)
	}
	if err != nil {
		return model.ModuleAccess{}, errors.WithStack(err)
	}
	return row.toDomainModel(), nil
}

func (r *moduleAccessRepository) FindByUserAndModule(userID, moduleID uuid.UUID) ([]model.ModuleAccess, error) {
	const query = `
		SELECT id, user_id, module_id, effect, source_id, created_at, updated_at
		FROM module_access
		WHERE user_id = ? AND module_id = ?
	`
	var rows []moduleAccessRow
	err := r.client.SelectContext(
		r.ctx,
		"find-module-access-by-user-and-module",
		&rows,
		query,
		newBinaryUUID(userID),
		newBinaryUUID(moduleID),
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := make([]model.ModuleAccess, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.toDomainModel())
	}
	return result, nil
}

func (r *moduleAccessRepository) Delete(id uuid.UUID) error {
	const query = `DELETE FROM module_access WHERE id = ?`
	_, err := r.client.ExecContext(r.ctx, "delete-module-access", query, newBinaryUUID(id))
	return errors.WithStack(err)
}

func (r *moduleAccessRepository) DeleteBySource(sourceID string) error {
	const query = `DELETE FROM module_access WHERE source_id=?`
	_, err := r.client.ExecContext(r.ctx, "delete-module-access-by-source", query, sourceID)
	return errors.WithStack(err)
}

func (r *moduleAccessRepository) DeleteBySourceAndUser(sourceID string, userID uuid.UUID) error {
	const query = `DELETE FROM module_access WHERE source_id=? AND user_id=?`
	_, err := r.client.ExecContext(
		r.ctx,
		"delete-module-access-by-source-user",
		query,
		sourceID,
		newBinaryUUID(userID),
	)
	return errors.WithStack(err)
}

func (r moduleAccessRow) toDomainModel() model.ModuleAccess {
	return model.ModuleAccess{
		ID:        r.ID.UUID(),
		UserID:    r.UserID.UUID(),
		ModuleID:  r.ModuleID.UUID(),
		Effect:    r.Effect,
		SourceID:  r.SourceID,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
