package repository

import (
	"context"
	"database/sql"
	"editor/internal/editor/domain/model"
	stderrors "errors"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"
)

type documentRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type documentRow struct {
	ID          uuid.UUID      `db:"id"`
	OwnerUserID uuid.UUID      `db:"owner_user_id"`
	WorkspaceID uuid.UUID      `db:"workspace_id"`
	Title       string         `db:"title"`
	Content     string         `db:"content"`
	Version     int64          `db:"version"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	Role        sql.NullString `db:"role"`
}

func NewDocument(ctx context.Context, client mysql.ClientContext) model.DocumentRepository {
	return &documentRepository{
		ctx:    ctx,
		client: client,
	}
}

func (*documentRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *documentRepository) Create(d model.Document) error {
	const query = `
		INSERT INTO document(
			 id,
			 owner_user_id,
			 workspace_id,
			 title,
			 content,
			 version,
			 created_at,
			 updated_at) 
		VALUES (?,?,?,?,?,?,?,?)
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"document.create",
		query,
		d.ID,
		d.OwnerUserID,
		d.WorkspaceID,
		d.Title,
		d.Content,
		d.Version,
		d.CreatedAt,
		d.UpdatedAt,
	)
	return errors.WithStack(err)
}

func (r *documentRepository) Find(id uuid.UUID) (model.Document, error) {
	const query = `
		SELECT 
		    id,
		    owner_user_id,
		    workspace_id,
		    title,
		    content,
		    version,
		    created_at,
		    updated_at
		FROM document 
		WHERE id=?
	`
	var row documentRow
	err := r.client.GetContext(r.ctx, "document.find", &row, query, id)
	if stderrors.Is(err, sql.ErrNoRows) {
		return model.Document{}, model.ErrDocumentNotFound
	}
	return convertToDomainModel(row), err
}

func (r *documentRepository) ListForUser(userID, workspaceID uuid.UUID) ([]model.Document, error) {
	const query = `
		SELECT 
		    d.id,
		    d.owner_user_id,
		    d.workspace_id,
		    d.title,
		    d.content,
		    d.version,
		    d.created_at,
		    d.updated_at,
			CASE 
				WHEN d.owner_user_id=? 
				THEN 'owner' 
				ELSE p.role 
			END role 
		FROM document d 
		LEFT JOIN document_permission p 
		    ON p.document_id=d.id AND p.user_id=? 
		WHERE d.workspace_id=? AND (d.owner_user_id=? OR p.user_id=?)
		ORDER BY d.updated_at DESC
	`
	var rows []documentRow
	err := r.client.SelectContext(
		r.ctx,
		"document.list",
		&rows,
		query,
		userID,
		userID,
		workspaceID,
		userID,
		userID,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := make([]model.Document, 0, len(rows))
	for _, row := range rows {
		result = append(result, convertToDomainModel(row))
	}
	return result, nil
}

func (r *documentRepository) Update(d model.Document, expected int64) error {
	const query = `
		UPDATE document 
		SET title=?,content=?,version=?,updated_at=? 
		WHERE id=? AND version=?
	`
	res, err := r.client.ExecContext(
		r.ctx,
		"document.update",
		query,
		d.Title,
		d.Content,
		d.Version,
		d.UpdatedAt,
		d.ID,
		expected,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrVersionConflict
	}
	return nil
}

func (r *documentRepository) Delete(id uuid.UUID) error {
	const query = `DELETE FROM document WHERE id=?`
	res, err := r.client.ExecContext(r.ctx, "document.delete", query, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrDocumentNotFound
	}
	return nil
}

func (r *documentRepository) FindRole(documentID, userID uuid.UUID) (model.Role, error) {
	const query = `SELECT role FROM document_permission WHERE document_id=? AND user_id=?`
	var role string
	err := r.client.GetContext(r.ctx, "permission.find", &role, query, documentID, userID)
	if stderrors.Is(err, sql.ErrNoRows) {
		return "", model.ErrDocumentNotFound
	}
	return model.Role(role), err
}

func (r *documentRepository) ListPermissions(documentID uuid.UUID) ([]model.Permission, error) {
	const query = `
		SELECT 
		    document_id,
		    user_id,
		    role,
		    created_at,
		    updated_at 
		FROM document_permission 
		WHERE document_id=? 
		ORDER BY created_at
	`
	var rows []struct {
		DocumentID uuid.UUID    `db:"document_id"`
		UserID     uuid.UUID    `db:"user_id"`
		Role       string       `db:"role"`
		CreatedAt  sql.NullTime `db:"created_at"`
		UpdatedAt  sql.NullTime `db:"updated_at"`
	}
	err := r.client.SelectContext(r.ctx, "permission.list", &rows, query, documentID)
	if err != nil {
		return nil, err
	}
	result := make([]model.Permission, 0, len(rows))
	for _, v := range rows {
		result = append(
			result,
			model.Permission{
				DocumentID: v.DocumentID,
				UserID:     v.UserID,
				Role:       model.Role(v.Role),
				CreatedAt:  v.CreatedAt.Time,
				UpdatedAt:  v.UpdatedAt.Time,
			})
	}
	return result, nil
}

func (r *documentRepository) StorePermission(p model.Permission) error {
	const query = `
		INSERT INTO document_permission
		    (document_id,user_id,role,created_at,updated_at) 
		VALUES(?,?,?,?,?) 
		ON DUPLICATE KEY UPDATE 
		    role=VALUES(role),
		 	updated_at=VALUES(updated_at)
	`
	_, err := r.client.ExecContext(
		r.ctx,
		"permission.store",
		query,
		p.DocumentID,
		p.UserID,
		p.Role,
		p.CreatedAt,
		p.UpdatedAt,
	)
	return err
}

func (r *documentRepository) DeletePermission(documentID, userID uuid.UUID) error {
	const query = `DELETE FROM document_permission WHERE document_id=? AND user_id=?`
	_, err := r.client.ExecContext(
		r.ctx,
		"permission.delete",
		query,
		documentID,
		userID,
	)
	return err
}

func convertToDomainModel(r documentRow) model.Document {
	role := model.Role(r.Role.String)
	return model.Document{
		ID:          r.ID,
		OwnerUserID: r.OwnerUserID,
		WorkspaceID: r.WorkspaceID,
		Title:       r.Title,
		Content:     r.Content,
		Version:     r.Version,
		CreatedAt:   r.CreatedAt.Time,
		UpdatedAt:   r.UpdatedAt.Time,
		Role:        role,
	}
}
