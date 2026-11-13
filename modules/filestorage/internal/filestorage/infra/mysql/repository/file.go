package repository

import (
	"context"
	"database/sql"
	"errors"
	"filestorage/internal/filestorage/domain"
	"time"
	"uuid"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type repo struct {
	ctx context.Context
	db  sdkmysql.ClientContext
}

func New(ctx context.Context, db sdkmysql.ClientContext) domain.Repository { return &repo{ctx, db} }

type row struct {
	ID             uuid.UUID `db:"id"`
	WorkspaceID    uuid.UUID `db:"workspace_id"`
	OwnerUserID    uuid.UUID `db:"owner_user_id"`
	ObjectKey      string    `db:"object_key"`
	FileName       string    `db:"file_name"`
	ContentType    string    `db:"content_type"`
	SizeBytes      int64     `db:"size_bytes"`
	ChecksumSHA256 string    `db:"checksum_sha256"`
	CreatedAt      time.Time `db:"created_at"`
}

func (r *repo) Store(v domain.File) error {
	_, e := r.db.ExecContext(r.ctx, "file.store", `INSERT INTO stored_file(id,workspace_id,owner_user_id,object_key,file_name,content_type,size_bytes,checksum_sha256,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.WorkspaceID, v.OwnerUserID, v.ObjectKey, v.FileName, v.ContentType, v.SizeBytes, v.ChecksumSHA256, v.CreatedAt)
	return e
}

func (r *repo) Find(id uuid.UUID) (domain.File, error) {
	var v row
	e := r.db.GetContext(r.ctx, "file.find", &v, `SELECT id,workspace_id,owner_user_id,object_key,file_name,content_type,size_bytes,checksum_sha256,created_at FROM stored_file WHERE id=?`, id)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return v.domain(), e
}

func (r *repo) List(w uuid.UUID) ([]domain.File, error) {
	var rows []row
	e := r.db.SelectContext(r.ctx, "file.list", &rows, `SELECT id,workspace_id,owner_user_id,object_key,file_name,content_type,size_bytes,checksum_sha256,created_at FROM stored_file WHERE workspace_id=? ORDER BY created_at DESC`, w)
	out := make([]domain.File, 0, len(rows))
	for _, v := range rows {
		out = append(out, v.domain())
	}
	return out, e
}

func (v row) domain() domain.File {
	return domain.File{ID: v.ID, WorkspaceID: v.WorkspaceID, OwnerUserID: v.OwnerUserID, ObjectKey: v.ObjectKey, FileName: v.FileName, ContentType: v.ContentType, SizeBytes: v.SizeBytes, ChecksumSHA256: v.ChecksumSHA256, CreatedAt: v.CreatedAt}
}

func (r *repo) Delete(id uuid.UUID) error {
	res, e := r.db.ExecContext(r.ctx, "file.delete", `DELETE FROM stored_file WHERE id=?`, id)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}
