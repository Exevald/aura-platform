package migrations

import (
	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

var migrations = []mysqlMigration{
	Version1791292600,
	Version1792830069,
	Version1793000010,
	Version1793100000,
}

type mysqlMigration func(client mysql.ClientContext) migrator.Migration

func Database(client mysql.ClientContext) []migrator.Migration {
	m := make([]migrator.Migration, 0, len(migrations))
	for _, buildFunc := range migrations {
		m = append(m, buildFunc(client))
	}
	return m
}
