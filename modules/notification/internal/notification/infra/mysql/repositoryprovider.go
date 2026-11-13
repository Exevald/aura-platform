package mysql

import (
	"context"
	"notification/internal/notification/domain"
	"notification/internal/notification/infra/mysql/repository"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type RepositoryProvider struct{ client sdkmysql.ClientContext }

func NewRepositoryProvider(c sdkmysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{c}
}

func (p *RepositoryProvider) Notification(ctx context.Context) domain.Repository {
	return repository.New(ctx, p.client)
}
