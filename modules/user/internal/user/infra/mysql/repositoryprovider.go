package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"user/internal/user/domain"
	"user/internal/user/infra/mysql/repository"
)

type RepositoryProvider struct {
	client mysql.ClientContext
}

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

func (p *RepositoryProvider) UserRepository(ctx context.Context) domain.UserRepository {
	return repository.NewUserRepository(ctx, p.client)
}
