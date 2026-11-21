package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"publication/internal/publication/app/service"
	"time"
	"uuid"

	"publication/internal/publication/app/model"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type ReleaseStore struct{ Client sdkmysql.ClientContext }

func (s *ReleaseStore) Get(ctx context.Context, id uuid.UUID) (model.Release, error) {
	const query = `
		SELECT id, owner_user_id, workspace_id, document_id, idempotency_key,
			expected_version, source_version, title, content, status, link_id,
			short_path, failure, delivery_status, created_at, published_at, revision
		FROM publication_release
		WHERE id = ?
	`
	var row releaseRow
	err := s.Client.GetContext(ctx, "publication.release.get", &row, query, id.String())
	if errors.Is(err, sql.ErrNoRows) {
		return model.Release{}, service.ErrNotFound
	}
	if err != nil {
		return model.Release{}, err
	}
	return row.toModel()
}

func (s *ReleaseStore) Store(ctx context.Context, r model.Release) (model.Release, error) {
	const query = `
		INSERT INTO publication_release (
			id, owner_user_id, workspace_id, document_id, idempotency_key,
			expected_version, source_version, title, content, status, link_id,
			short_path, failure, delivery_status, created_at, published_at, revision
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			source_version = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(source_version), source_version),
			title = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(title), title),
			content = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(content), content),
			status = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(status), status),
			link_id = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(link_id), link_id),
			short_path = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(short_path), short_path),
			failure = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(failure), failure),
			delivery_status = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(delivery_status), delivery_status),
			published_at = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(published_at), published_at),
			revision = IF(id = VALUES(id) AND revision + 1 = VALUES(revision), VALUES(revision), revision)
	`
	var published any
	if !r.PublishedAt.IsZero() {
		published = r.PublishedAt
	}
	linkID := ""
	if r.LinkID != uuid.Nil() {
		linkID = r.LinkID.String()
	}
	result, err := s.Client.ExecContext(
		ctx,
		"publication.release.store",
		query,
		r.ID.String(),
		r.OwnerID.String(),
		r.WorkspaceID.String(),
		r.DocumentID.String(),
		r.IdempotencyKey,
		r.ExpectedVersion,
		r.SourceVersion,
		r.Title,
		r.Content,
		r.Status,
		linkID,
		r.ShortPath,
		r.Failure,
		defaultDeliveryStatus(r.DeliveryStatus),
		createdAt(r.CreatedAt),
		published,
		r.Revision+1,
	)
	if err != nil {
		return model.Release{}, err
	}
	if r.Revision > 0 {
		rows, err2 := result.RowsAffected()
		if err2 != nil {
			return model.Release{}, err2
		}
		if rows == 0 {
			return model.Release{}, service.ErrConflict
		}
	}
	if r.Revision == 0 {
		return s.byIdempotency(ctx, r.OwnerID, r.IdempotencyKey)
	}
	stored, err := s.Get(ctx, r.ID)
	if err != nil {
		return model.Release{}, err
	}
	if stored.Revision != r.Revision+1 {
		return model.Release{}, service.ErrConflict
	}
	return stored, nil
}

func (s *ReleaseStore) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM publication_release WHERE id = ?`
	_, err := s.Client.ExecContext(ctx, "publication.release.delete", query, id.String())
	return err
}

func (s *ReleaseStore) byIdempotency(ctx context.Context, owner uuid.UUID, key string) (model.Release, error) {
	const query = `
		SELECT id, owner_user_id, workspace_id, document_id, idempotency_key,
			expected_version, source_version, title, content, status, link_id,
			short_path, failure, delivery_status, created_at, published_at, revision
		FROM publication_release
		WHERE owner_user_id = ? AND idempotency_key = ?
	`
	var row releaseRow
	err := s.Client.GetContext(ctx, "publication.release.idempotency", &row, query, owner.String(), key)
	if err != nil {
		return model.Release{}, err
	}
	return row.toModel()
}

type releaseRow struct {
	ID              uuid.UUID    `db:"id"`
	OwnerID         uuid.UUID    `db:"owner_user_id"`
	WorkspaceID     uuid.UUID    `db:"workspace_id"`
	DocumentID      uuid.UUID    `db:"document_id"`
	IdempotencyKey  string       `db:"idempotency_key"`
	ExpectedVersion int64        `db:"expected_version"`
	SourceVersion   int64        `db:"source_version"`
	Title           string       `db:"title"`
	Content         string       `db:"content"`
	Status          string       `db:"status"`
	LinkID          string       `db:"link_id"`
	ShortPath       string       `db:"short_path"`
	Failure         string       `db:"failure"`
	DeliveryStatus  string       `db:"delivery_status"`
	CreatedAt       time.Time    `db:"created_at"`
	PublishedAt     sql.NullTime `db:"published_at"`
	Revision        int64        `db:"revision"`
}

func (r releaseRow) toModel() (model.Release, error) {
	linkID := uuid.Nil()
	if r.LinkID != "" {
		var err error
		linkID, err = uuid.Parse(r.LinkID)
		if err != nil {
			return model.Release{}, fmt.Errorf("invalid stored link ID %q: %w", r.LinkID, err)
		}
	}
	v := model.Release{
		ID: r.ID, OwnerID: r.OwnerID, WorkspaceID: r.WorkspaceID, DocumentID: r.DocumentID,
		IdempotencyKey: r.IdempotencyKey, ExpectedVersion: r.ExpectedVersion, SourceVersion: r.SourceVersion,
		Title: r.Title, Content: r.Content, Status: r.Status, LinkID: linkID, ShortPath: r.ShortPath,
		Failure: r.Failure, DeliveryStatus: r.DeliveryStatus, CreatedAt: r.CreatedAt, Revision: r.Revision,
	}
	if r.PublishedAt.Valid {
		v.PublishedAt = r.PublishedAt.Time
	}
	return v, nil
}

func defaultDeliveryStatus(status string) string {
	if status == "" {
		return "pending"
	}
	return status
}

func createdAt(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value
}
