package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"access/internal/access/domain"
	"access/internal/access/infra/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

type RepositoryProvider struct {
	client mysql.ClientContext
}

func (p *RepositoryProvider) ModuleAccess(ctx context.Context) domain.ModuleAccessRepository {
	return repository.NewModuleAccessRepository(ctx, p.client)
}
