package repository

import (
	"context"
	"database/sql"
	stderrors "errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"

	"access/internal/access/domain"
)

type moduleAccessRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func NewModuleAccessRepository(
	ctx context.Context,
	client mysql.ClientContext,
) domain.ModuleAccessRepository {
	return &moduleAccessRepository{
		ctx:    ctx,
		client: client,
	}
}

func (r *moduleAccessRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *moduleAccessRepository) Store(access domain.ModuleAccess) error {
	const query = `
		INSERT INTO module_access
			(id, user_id, module_id, effect, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			module_id = VALUES(module_id),
			effect = VALUES(effect),
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
		access.CreatedAt,
		access.UpdatedAt,
	)
	return errors.WithStack(err)
}

func (r *moduleAccessRepository) Find(id uuid.UUID) (domain.ModuleAccess, error) {
	const query = `
		SELECT id, user_id, module_id, effect, created_at, updated_at
		FROM module_access
		WHERE id = ?
	`
	var row moduleAccessRow
	err := r.client.GetContext(r.ctx, "find-module-access", &row, query, newBinaryUUID(id))
	if stderrors.Is(err, sql.ErrNoRows) {
		return domain.ModuleAccess{}, errors.WithStack(domain.ErrModuleAccessNotFound)
	}
	if err != nil {
		return domain.ModuleAccess{}, errors.WithStack(err)
	}
	return row.toDomainModel(), nil
}

func (r *moduleAccessRepository) FindByUserAndModule(userID, moduleID uuid.UUID) ([]domain.ModuleAccess, error) {
	const query = `
		SELECT id, user_id, module_id, effect, created_at, updated_at
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
	result := make([]domain.ModuleAccess, 0, len(rows))
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

type moduleAccessRow struct {
	ID        binaryUUID    `db:"id"`
	UserID    binaryUUID    `db:"user_id"`
	ModuleID  binaryUUID    `db:"module_id"`
	Effect    domain.Effect `db:"effect"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

func (r moduleAccessRow) toDomainModel() domain.ModuleAccess {
	return domain.ModuleAccess{
		ID:        r.ID.UUID(),
		UserID:    r.UserID.UUID(),
		ModuleID:  r.ModuleID.UUID(),
		Effect:    r.Effect,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
