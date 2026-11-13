package repository

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"usergroup/internal/usergroup/domain"
)

type groupMembershipRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type groupMembershipRow struct {
	GroupID   uuid.UUID `db:"group_id"`
	UserID    uuid.UUID `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

func NewGroupMembership(ctx context.Context, client mysql.ClientContext) domain.GroupMembershipRepository {
	return &groupMembershipRepository{ctx: ctx, client: client}
}

func (r *groupMembershipRepository) ListByGroup(groupID uuid.UUID) ([]domain.GroupMembership, error) {
	const query = `
		SELECT group_id, user_id, created_at
		FROM user_group_member
		WHERE group_id=?
		ORDER BY created_at
	`
	var rows []groupMembershipRow
	if err := r.client.SelectContext(r.ctx, "group-membership.list", &rows, query, groupID); err != nil {
		return nil, err
	}
	result := make([]domain.GroupMembership, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.domain())
	}
	return result, nil
}

func (r *groupMembershipRepository) Store(membership domain.GroupMembership) (bool, error) {
	const query = `
		INSERT IGNORE INTO user_group_member(group_id, user_id, created_at)
		VALUES(?, ?, ?)
	`
	result, err := r.client.ExecContext(
		r.ctx,
		"group-membership.store",
		query,
		membership.GroupID,
		membership.UserID,
		membership.CreatedAt,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (r *groupMembershipRepository) Delete(groupID, userID uuid.UUID) (bool, error) {
	const query = `DELETE FROM user_group_member WHERE group_id=? AND user_id=?`
	result, err := r.client.ExecContext(
		r.ctx,
		"group-membership.delete",
		query,
		groupID,
		userID,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	return affected > 0, err
}

func (r groupMembershipRow) domain() domain.GroupMembership {
	return domain.GroupMembership{
		GroupID:   r.GroupID,
		UserID:    r.UserID,
		CreatedAt: r.CreatedAt,
	}
}
