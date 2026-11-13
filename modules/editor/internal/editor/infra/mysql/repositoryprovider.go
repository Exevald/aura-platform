package mysql

import (
	"context"
	"editor/internal/editor/domain/model"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"editor/internal/editor/infra/mysql/repository"
)

type RepositoryProvider struct {
	client mysql.ClientContext
}

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

func (p *RepositoryProvider) Document(ctx context.Context) model.DocumentRepository {
	return repository.NewDocument(ctx, p.client)
}

func (p *RepositoryProvider) ModuleShare(ctx context.Context) model.ModuleShareRepository {
	return repository.NewShare(ctx, p.client)
}

func (p *RepositoryProvider) Attachment(ctx context.Context) model.AttachmentRepository {
	return repository.NewAttachment(ctx, p.client)
}
