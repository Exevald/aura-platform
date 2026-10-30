package app

import (
	"context"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/uow"
	domainevent "github.com/distributed-programming-2026/lib/event"

	"editor/internal/editor/domain"
	"editor/internal/editor/infra/mysql"
)

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type User struct {
	ID                 uuid.UUID
	Email, DisplayName string
}

type UserClient interface {
	FindByEmail(context.Context, string) (User, error)
	Get(context.Context, uuid.UUID) (User, error)
}

type Permission struct {
	UserID             uuid.UUID
	Email, DisplayName string
	Role               domain.Role
	UpdatedAt          time.Time
}

type EditorService interface {
	Create(ctx context.Context, actor uuid.UUID, title, content string) (out domain.Document, err error)
	Get(ctx context.Context, actor, id uuid.UUID) (out domain.Document, err error)
	List(ctx context.Context, actor uuid.UUID) (out []domain.Document, err error)
	Update(
		ctx context.Context,
		actor,
		id uuid.UUID,
		title,
		content string,
		version int64,
	) (out domain.Document, err error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Grant(ctx context.Context, userID, id uuid.UUID, email string, role domain.Role) error
	Revoke(ctx context.Context, actor, id, userID uuid.UUID) error
	Permissions(ctx context.Context, actor, id uuid.UUID) ([]Permission, error)
}

type service struct {
	unit        uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider]
	dispatch    DispatcherFactory
	userAdapter UserClient
}

func NewService(
	unit uow.UnitOfWorkWithRepositoryProvider[*mysql.RepositoryProvider],
	dispatch DispatcherFactory,
	userAdapter UserClient,
) EditorService {
	return &service{
		unit:        unit,
		dispatch:    dispatch,
		userAdapter: userAdapter,
	}
}

func (s *service) Create(
	ctx context.Context,
	actor uuid.UUID,
	title,
	content string,
) (document domain.Document, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		document, err = s.domainService(ctx, p).Create(actor, title, content)
		return err
	})
	return
}

func (s *service) Get(ctx context.Context, actor, id uuid.UUID) (document domain.Document, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		document, err = s.domainService(ctx, p).Get(actor, id)
		return err
	})
	return
}

func (s *service) List(ctx context.Context, actor uuid.UUID) (documents []domain.Document, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		documents, err = s.domainService(ctx, p).List(actor)
		return err
	})
	return
}

func (s *service) Update(
	ctx context.Context,
	actor,
	id uuid.UUID,
	title,
	content string,
	version int64,
) (document domain.Document, err error) {
	err = s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		document, err = s.domainService(ctx, p).Update(actor, id, title, content, version)
		return err
	})
	return
}

func (s *service) Delete(ctx context.Context, actor, id uuid.UUID) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		return s.domainService(ctx, p).Delete(actor, id)
	})
}

func (s *service) Grant(ctx context.Context, actor, id uuid.UUID, email string, role domain.Role) error {
	user, err := s.userAdapter.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		return s.domainService(ctx, p).Grant(actor, id, user.ID, role)
	})
}
func (s *service) Revoke(ctx context.Context, actor, id, userID uuid.UUID) error {
	return s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		return s.domainService(ctx, p).Revoke(actor, id, userID)
	})
}

func (s *service) Permissions(ctx context.Context, actor, id uuid.UUID) ([]Permission, error) {
	var values []domain.Permission
	err := s.unit.ExecuteWithRepositoryProvider(ctx, func(p *mysql.RepositoryProvider) error {
		var e error
		values, e = s.domainService(ctx, p).Permissions(actor, id)
		return e
	})
	if err != nil {
		return nil, err
	}
	result := make([]Permission, 0, len(values))
	for _, v := range values {
		user, e := s.userAdapter.Get(ctx, v.UserID)
		if e != nil {
			return nil, e
		}
		result = append(
			result,
			Permission{
				UserID:      v.UserID,
				Email:       user.Email,
				DisplayName: user.DisplayName,
				Role:        v.Role,
				UpdatedAt:   v.UpdatedAt,
			})
	}
	return result, nil
}

func (s *service) domainService(ctx context.Context, p *mysql.RepositoryProvider) domain.DocumentService {
	return domain.NewService(p.Document(ctx), s.dispatch.NewDispatcher(ctx))
}
