package mysql

import (
	"context"
	"workspace/internal/workspace/domain"
	"workspace/internal/workspace/infra/mysql/repository"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type RepositoryProvider struct{ client sdkmysql.ClientContext }

func NewRepositoryProvider(client sdkmysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{client: client}
}

func (p *RepositoryProvider) Workspace(ctx context.Context) domain.Repository {
	return repository.New(ctx, p.client)
}
