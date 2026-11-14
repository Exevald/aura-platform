package repository

import (
	"context"
	"database/sql"
	stderrors "errors"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/pkg/errors"

	"user/internal/user/domain"
)

type userRepository struct {
	ctx    context.Context
	client mysql.ClientContext
}

type sqlxUser struct {
	ID          uuid.UUID `db:"id"`
	DisplayName string    `db:"display_name"`
	FirstName   string    `db:"first_name"`
	LastName    string    `db:"last_name"`
	Email       string    `db:"email"`
}

func NewUserRepository(
	ctx context.Context,
	client mysql.ClientContext,
) domain.UserRepository {
	return &userRepository{
		ctx:    ctx,
		client: client,
	}
}

func (u *userRepository) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (u *userRepository) Store(user domain.User) error {
	const query = `
		INSERT INTO user (
		    id,
		    display_name,
		    first_name,
		    last_name,
		    email
		)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    display_name = VALUES (display_name),
		    first_name = VALUES (first_name),
		    last_name = VALUES (last_name),
		    email = VALUES (email)
	`
	_, err := u.client.ExecContext(
		u.ctx,
		"insert-into-user",
		query,
		user.ID,
		user.DisplayName,
		user.FirstName,
		user.LastName,
		user.Email,
	)
	return errors.WithStack(err)
}

func (u *userRepository) Find(id uuid.UUID) (domain.User, error) {
	const query = `
		SELECT
		    id,
		    display_name,
		    first_name,
		    last_name,
		    email
		FROM user
		WHERE id = ? AND deleted_at IS NULL
	`
	var user sqlxUser
	err := u.client.GetContext(u.ctx, "select-user", &user, query, id)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.WithStack(domain.ErrUserNotFound)
		}
		return domain.User{}, errors.WithStack(err)
	}

	return domain.User{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
	}, nil
}

func (u *userRepository) FindByEmail(email string) (domain.User, error) {
	const query = `
		SELECT
		    id,
		    display_name,
		    first_name,
		    last_name,
		    email
		FROM user
		WHERE email = ? AND deleted_at IS NULL
	`
	var user sqlxUser
	err := u.client.GetContext(u.ctx, "select-user-by-email", &user, query, email)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.WithStack(domain.ErrUserNotFound)
		}
		return domain.User{}, errors.WithStack(err)
	}
	return domain.User{ID: user.ID, DisplayName: user.DisplayName, FirstName: user.FirstName, LastName: user.LastName, Email: user.Email}, nil
}

func (u *userRepository) Delete(id uuid.UUID) error {
	const query = `
		UPDATE user
		SET
		    deleted_at = NOW()
		WHERE id = ? AND deleted_at is NULL
	`
	_, err := u.client.ExecContext(u.ctx, "remove-user", query, (id))
	return errors.WithStack(err)
}
