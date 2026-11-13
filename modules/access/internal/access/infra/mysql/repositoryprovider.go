package mysql

import (
	"access/internal/access/domain/model"
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"access/internal/access/infra/mysql/repository"
)

type RepositoryProvider struct {
	client mysql.ClientContext
}

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

func (p *RepositoryProvider) ModuleAccess(ctx context.Context) model.ModuleAccessRepository {
	return repository.NewModuleAccessRepository(ctx, p.client)
}

func (p *RepositoryProvider) ModuleShareOperation(ctx context.Context) model.ModuleShareRepository {
	return repository.NewModuleShareOperation(ctx, p.client)
}
