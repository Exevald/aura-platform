package mysql

import "github.com/distributed-programming-2026/go-sdk/pkg/mysql"

type RepositoryProvider struct{}

func NewRepositoryProvider(_ mysql.ClientContext) *RepositoryProvider {
	return &RepositoryProvider{}
}
