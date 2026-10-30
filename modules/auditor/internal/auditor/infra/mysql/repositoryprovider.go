package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"auditor/internal/auditor/domain"
	"auditor/internal/auditor/infra/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

type RepositoryProvider struct {
	client mysql.ClientContext
}

func (p *RepositoryProvider) Audit(ctx context.Context) domain.AuditRepository {
	return repository.NewAuditRepository(ctx, p.client)
}
