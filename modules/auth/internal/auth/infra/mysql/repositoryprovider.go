package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"auth/internal/auth/domain"
	"auth/internal/auth/infra/mysql/repository"
)

type RepositoryProvider struct {
	client mysql.ClientContext
}

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

func (p *RepositoryProvider) Credential(ctx context.Context) domain.CredentialRepository {
	return repository.NewCredentialRepository(ctx, p.client)
}
