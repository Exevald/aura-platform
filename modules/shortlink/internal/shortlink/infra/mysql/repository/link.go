package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	sqldriver "github.com/go-sql-driver/mysql"

	"shortlink/internal/shortlink/domain"
)

type linkRepository struct {
	ctx context.Context
	db  mysql.ClientContext
}

type linkRow struct {
	ID        string    `db:"id"`
	OwnerID   string    `db:"owner_user_id"`
	Code      string    `db:"code"`
	TargetURL string    `db:"target_url"`
	CreatedAt time.Time `db:"created_at"`
}

func NewLink(ctx context.Context, db mysql.ClientContext) domain.LinkRepository {
	return &linkRepository{ctx: ctx, db: db}
}

func (r *linkRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *linkRepository) Store(link domain.Link) error {
	_, err := r.db.ExecContext(r.ctx, "shortlink.store", `INSERT INTO short_link(id,owner_user_id,code,target_url,created_at) VALUES(?,?,?,?,?)`, link.ID.String(), link.OwnerID.String(), link.Code, link.TargetURL, link.CreatedAt)
	var mysqlErr *sqldriver.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "code") {
		return domain.ErrCodeConflict
	}
	return err
}

func (r *linkRepository) ListByOwner(ownerID uuid.UUID) ([]domain.Link, error) {
	var rows []linkRow
	if err := r.db.SelectContext(r.ctx, "shortlink.list", &rows, `SELECT id,owner_user_id,code,target_url,created_at FROM short_link WHERE owner_user_id=? ORDER BY created_at DESC`, ownerID.String()); err != nil {
		return nil, err
	}
	result := make([]domain.Link, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.domain())
	}
	return result, nil
}

func (r *linkRepository) FindByCode(code string) (domain.Link, error) {
	var row linkRow
	err := r.db.GetContext(r.ctx, "shortlink.find-code", &row, `SELECT id,owner_user_id,code,target_url,created_at FROM short_link WHERE code=?`, code)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Link{}, domain.ErrNotFound
	}
	return row.domain(), err
}

func (r *linkRepository) Delete(ownerID, id uuid.UUID) error {
	result, err := r.db.ExecContext(r.ctx, "shortlink.delete", `DELETE FROM short_link WHERE id=? AND owner_user_id=?`, id.String(), ownerID.String())
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r linkRow) domain() domain.Link {
	return domain.Link{
		ID:        uuid.MustParse(r.ID),
		OwnerID:   uuid.MustParse(r.OwnerID),
		Code:      r.Code,
		TargetURL: r.TargetURL,
		CreatedAt: r.CreatedAt,
	}
}
