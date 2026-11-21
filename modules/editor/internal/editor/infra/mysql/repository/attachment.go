package repository

import (
	"context"
	"database/sql"
	"editor/internal/editor/domain/model"
	stderrors "errors"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type attachmentRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

func NewAttachment(ctx context.Context, c mysql.ClientContext) model.AttachmentRepository {
	return &attachmentRepository{ctx, c}
}

type attachmentRow struct {
	DocumentID  uuid.UUID `db:"document_id"`
	FileID      uuid.UUID `db:"file_id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	FileName    string    `db:"file_name"`
	ContentType string    `db:"content_type"`
	SizeBytes   int64     `db:"size_bytes"`
	CreatedAt   time.Time `db:"created_at"`
}

func (r *attachmentRepository) StoreAttachment(v model.Attachment) error {
	const sqlAttachmentStoreQuery = `
		INSERT INTO document_attachment(
		    document_id,
		    file_id,
		    workspace_id,
		    file_name,
		    content_type,
		    size_bytes,
		    created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, e := r.client.ExecContext(r.ctx, "attachment.store", sqlAttachmentStoreQuery, v.DocumentID, v.FileID.String(), v.WorkspaceID, v.FileName, v.ContentType, v.SizeBytes, v.CreatedAt)
	return e
}

func (r *attachmentRepository) ListAttachments(id uuid.UUID) ([]model.Attachment, error) {
	const sqlAttachmentListQuery = `
		SELECT
		    document_id,
		    file_id,
		    workspace_id,
		    file_name,
		    content_type,
		    size_bytes,
		    created_at
		FROM document_attachment
		WHERE document_id = ?
		ORDER BY created_at
	`

	var rows []attachmentRow
	e := r.client.SelectContext(r.ctx, "attachment.list", &rows, sqlAttachmentListQuery, id)
	out := make([]model.Attachment, 0, len(rows))
	for _, v := range rows {
		out = append(out, model.Attachment(v))
	}
	return out, e
}

func (r *attachmentRepository) FindAttachment(doc, file uuid.UUID) (model.Attachment, error) {
	const sqlAttachmentFindQuery = `
		SELECT
		    document_id,
		    file_id,
		    workspace_id,
		    file_name,
		    content_type,
		    size_bytes,
		    created_at
		FROM document_attachment
		WHERE document_id = ? AND file_id = ?
	`

	var v attachmentRow
	e := r.client.GetContext(r.ctx, "attachment.find", &v, sqlAttachmentFindQuery, doc, file.String())
	if stderrors.Is(e, sql.ErrNoRows) {
		e = model.ErrDocumentNotFound
	}
	return model.Attachment(v), e
}

func (r *attachmentRepository) DeleteAttachment(doc, file uuid.UUID) error {
	const sqlAttachmentDeleteQuery = `
		DELETE FROM document_attachment
		WHERE document_id = ? AND file_id = ?
	`

	res, e := r.client.ExecContext(r.ctx, "attachment.delete", sqlAttachmentDeleteQuery, doc, file.String())
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return model.ErrDocumentNotFound
		}
	}
	return e
}
