package storage

import (
	"context"
	"publication/internal/publication/app/model"
	"uuid"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type ReleaseQueries struct{ Client sdkmysql.ClientContext }

type AttachmentQueries struct{ Client sdkmysql.ClientContext }

func (q *ReleaseQueries) List(ctx context.Context, owner, workspace uuid.UUID) ([]model.Release, error) {
	const query = `
		SELECT id, owner_user_id, workspace_id, document_id, idempotency_key,
			expected_version, source_version, title, content, status, link_id,
			short_path, failure, delivery_status, created_at, published_at, revision
		FROM publication_release
		WHERE owner_user_id = ? AND workspace_id = ?
		ORDER BY created_at DESC
	`
	var rows []releaseRow
	err := q.Client.SelectContext(ctx, "publication.release.list", &rows, query, owner.String(), workspace.String())
	if err != nil {
		return nil, err
	}
	return releases(rows)
}

func (q *ReleaseQueries) ByDocument(ctx context.Context, document uuid.UUID) ([]model.Release, error) {
	const query = `
		SELECT id, owner_user_id, workspace_id, document_id, idempotency_key,
			expected_version, source_version, title, content, status, link_id,
			short_path, failure, delivery_status, created_at, published_at, revision
		FROM publication_release
		WHERE document_id = ? AND status IN ('pending', 'active', 'closing')
	`
	var rows []releaseRow
	err := q.Client.SelectContext(ctx, "publication.release.document", &rows, query, document.String())
	if err != nil {
		return nil, err
	}
	return releases(rows)
}

func (q *ReleaseQueries) Pending(ctx context.Context) ([]model.Release, error) {
	const query = `
		SELECT id, owner_user_id, workspace_id, document_id, idempotency_key,
			expected_version, source_version, title, content, status, link_id,
			short_path, failure, delivery_status, created_at, published_at, revision
		FROM publication_release
		WHERE status IN ('pending', 'closing')
		ORDER BY created_at LIMIT 100
	`
	var rows []releaseRow
	err := q.Client.SelectContext(ctx, "publication.release.pending", &rows, query)
	if err != nil {
		return nil, err
	}
	return releases(rows)
}

func (q *AttachmentQueries) List(ctx context.Context, release uuid.UUID) ([]model.Attachment, error) {
	const query = `
		SELECT release_id, file_id, copy_file_id, file_name, content_type, size_bytes
		FROM publication_attachment
		WHERE release_id = ?
		ORDER BY file_name
	`
	var rows []attachmentRow
	err := q.Client.SelectContext(ctx, "publication.attachment.list", &rows, query, release.String())
	out := make([]model.Attachment, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toDomainModel())
	}
	return out, err
}

func releases(rows []releaseRow) ([]model.Release, error) {
	out := make([]model.Release, 0, len(rows))
	for _, row := range rows {
		release, err := row.toModel()
		if err != nil {
			return nil, err
		}
		out = append(out, release)
	}
	return out, nil
}
