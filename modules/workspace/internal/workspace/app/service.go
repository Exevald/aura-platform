package app

import (
	"context"
	"uuid"
	"workspace/internal/workspace/domain"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	Workspace(context.Context) domain.Repository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type Service struct {
	unit   UnitOfWork
	events DispatcherFactory
}

func NewService(unit UnitOfWork, events DispatcherFactory) *Service {
	return &Service{unit: unit, events: events}
}

func (s *Service) run(ctx context.Context, fn func(*domain.Service) error) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return fn(domain.NewService(p.Workspace(ctx), s.events.NewDispatcher(ctx)))
	})
}

func (s *Service) Create(ctx context.Context, user uuid.UUID, name string) (out domain.Workspace, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateTeam(user, name)
		return err
	})
	return
}

func (s *Service) EnsurePersonal(ctx context.Context, id, user uuid.UUID, name string) (out domain.Workspace, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.EnsurePersonal(id, user, name)
		return err
	})
	return
}

func (s *Service) List(ctx context.Context, user uuid.UUID) (out []domain.Workspace, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.List(user)
		return err
	})
	return
}

func (s *Service) Get(ctx context.Context, user, id uuid.UUID) (w domain.Workspace, m []domain.Member, r domain.Role, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		w, m, r, err = d.Get(user, id)
		return err
	})
	return
}

func (s *Service) Check(ctx context.Context, user, id uuid.UUID) (m domain.Member, w domain.Workspace, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		m, w, err = d.Check(user, id)
		return err
	})
	return
}

func (s *Service) PutMember(ctx context.Context, a, id, u uuid.UUID, r domain.Role) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.PutMember(a, id, u, r)
	})
}

func (s *Service) RemoveMember(ctx context.Context, a, id, u uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.RemoveMember(a, id, u)
	})
}

func (s *Service) Delete(ctx context.Context, a, id uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.Delete(a, id)
	})
}

func (s *Service) CreateInvitation(ctx context.Context, a, w, u uuid.UUID, email string, role domain.Role) (out domain.Invitation, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateInvitation(a, w, u, email, role)
		return err
	})
	return
}

func (s *Service) ListInvitations(ctx context.Context, a, w uuid.UUID) (out []domain.Invitation, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.ListInvitations(a, w)
		return err
	})
	return
}

func (s *Service) AcceptInvitation(ctx context.Context, a uuid.UUID, token string) (out domain.Workspace, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.AcceptInvitation(a, token)
		return err
	})
	return
}

func (s *Service) RevokeInvitation(ctx context.Context, a, w, id uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.RevokeInvitation(a, w, id)
	})
}
