package mysql

import (
	"calendar/internal/calendar/domain"
	"calendar/internal/calendar/infra/mysql/repository"
	"context"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type RepositoryProvider struct{ client sdkmysql.ClientContext }

func NewRepositoryProvider(c sdkmysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{c}
}

func (p *RepositoryProvider) Calendar(ctx context.Context) domain.Repository {
	return repository.New(ctx, p.client)
}
