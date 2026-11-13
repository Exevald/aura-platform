package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"session/internal/session/domain"
	"session/internal/session/infra/mysql/repository"
)

type RepositoryProvider struct {
	client mysql.ClientContext
}

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

func (p *RepositoryProvider) Session(ctx context.Context) domain.SessionRepository {
	return repository.NewSessionRepository(ctx, p.client)
}
