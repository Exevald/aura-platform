package app

import (
	"context"
	"time"
	"uuid"

	"shortlink/internal/shortlink/domain"
)

type RepositoryProvider interface {
	Link(context.Context) domain.LinkRepository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type Service struct {
	unit UnitOfWork
}

func NewService(unit UnitOfWork) *Service {
	return &Service{unit: unit}
}

func (s *Service) Create(ctx context.Context, ownerID uuid.UUID, target string) (link domain.Link, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		link, err = domain.NewLinkService(provider.Link(ctx)).Create(ownerID, target, time.Now().UTC())
		return err
	})
	return
}

func (s *Service) List(ctx context.Context, ownerID uuid.UUID) (links []domain.Link, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		links, err = domain.NewLinkService(provider.Link(ctx)).List(ownerID)
		return err
	})
	return
}

func (s *Service) Resolve(ctx context.Context, code string) (link domain.Link, err error) {
	err = s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		link, err = domain.NewLinkService(provider.Link(ctx)).Resolve(code)
		return err
	})
	return
}

func (s *Service) Delete(ctx context.Context, ownerID, id uuid.UUID) error {
	return s.unit.Execute(ctx, func(provider RepositoryProvider) error {
		return domain.NewLinkService(provider.Link(ctx)).Delete(ownerID, id)
	})
}
