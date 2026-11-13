package mysql

import "github.com/distributed-programming-2026/go-sdk/pkg/mysql"

func NewRepositoryProvider(client mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{
		client: client,
	}
}

type RepositoryProvider struct {
	client mysql.ClientContext
}
