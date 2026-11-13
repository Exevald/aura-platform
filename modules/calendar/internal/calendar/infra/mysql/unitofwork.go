package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/uow"

	"calendar/internal/calendar/app"
)

type UnitOfWork struct {
	unit uow.UnitOfWorkWithRepositoryProvider[*RepositoryProvider]
}

func NewUnitOfWork(unit uow.UnitOfWorkWithRepositoryProvider[*RepositoryProvider]) *UnitOfWork {
	return &UnitOfWork{unit: unit}
}

func (u *UnitOfWork) Execute(ctx context.Context, fn func(app.RepositoryProvider) error) error {
	return u.unit.ExecuteWithRepositoryProvider(ctx, func(provider *RepositoryProvider) error {
		return fn(provider)
	})
}
