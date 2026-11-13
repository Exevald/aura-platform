package mysql

import (
	"context"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"

	"registerer/internal/registerer/app"
	"registerer/internal/registerer/infra/mysql/storage"
)

type StorageProvider struct {
	client mysql.ClientContext
}

func NewStorageProvider(client mysql.ClientContext) *StorageProvider {
	return &StorageProvider{
		client: client,
	}
}

func (p *StorageProvider) Registration(ctx context.Context) app.RegistrationStorage {
	return storage.NewRegistrationStorage(ctx, p.client)
}
