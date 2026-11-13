package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"shortlink/internal/shortlink/domain"
	"shortlink/internal/shortlink/infra/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

type RepositoryProvider struct {
	client mysql.ClientContext
}

func (p *RepositoryProvider) Link(ctx context.Context) domain.LinkRepository {
	return repository.NewLink(ctx, p.client)
}
