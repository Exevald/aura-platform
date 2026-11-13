package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"usergroup/internal/usergroup/domain"
	"usergroup/internal/usergroup/infra/mysql/repository"
)

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

type RepositoryProvider struct {
	client mysql.ClientContext
}

func (p *RepositoryProvider) Group(ctx context.Context) domain.GroupRepository {
	return repository.NewGroup(ctx, p.client)
}

func (p *RepositoryProvider) GroupMembership(ctx context.Context) domain.GroupMembershipRepository {
	return repository.NewGroupMembership(ctx, p.client)
}
