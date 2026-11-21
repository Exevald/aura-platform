package storage

import (
	"context"
	"database/sql"
	"errors"
	"publication/internal/publication/app/service"
	"uuid"

	"publication/internal/publication/app/model"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type AttachmentStore struct{ Client sdkmysql.ClientContext }

func (s *AttachmentStore) Get(ctx context.Context, release, file uuid.UUID) (model.Attachment, error) {
	const query = `
		SELECT release_id, file_id, copy_file_id, file_name, content_type, size_bytes
		FROM publication_attachment
		WHERE release_id = ? AND file_id = ?
	`
	var row attachmentRow
	err := s.Client.GetContext(ctx, "publication.attachment.get", &row, query, release.String(), file.String())
	if errors.Is(err, sql.ErrNoRows) {
		return model.Attachment{}, service.ErrNotFound
	}
	return row.toDomainModel(), err
}

func (s *AttachmentStore) Store(ctx context.Context, a model.Attachment) error {
	const query = `
		INSERT INTO publication_attachment (
			release_id, file_id, copy_file_id, file_name, content_type, size_bytes
		) VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			copy_file_id = VALUES(copy_file_id),
			file_name = VALUES(file_name),
			content_type = VALUES(content_type),
			size_bytes = VALUES(size_bytes)
	`
	_, err := s.Client.ExecContext(ctx, "publication.attachment.store", query,
		a.ReleaseID.String(), a.FileID.String(), a.CopyID.String(), a.FileName, a.ContentType, a.SizeBytes)
	return err
}

func (s *AttachmentStore) Delete(ctx context.Context, release uuid.UUID) error {
	const query = `DELETE FROM publication_attachment WHERE release_id = ?`
	_, err := s.Client.ExecContext(ctx, "publication.attachment.delete", query, release.String())
	return err
}

type attachmentRow struct {
	ReleaseID   uuid.UUID `db:"release_id"`
	FileID      uuid.UUID `db:"file_id"`
	CopyID      uuid.UUID `db:"copy_file_id"`
	FileName    string    `db:"file_name"`
	ContentType string    `db:"content_type"`
	SizeBytes   int64     `db:"size_bytes"`
}

func (r attachmentRow) toDomainModel() model.Attachment {
	return model.Attachment(r)
}
