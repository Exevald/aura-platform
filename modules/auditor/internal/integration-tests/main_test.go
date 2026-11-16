package integration_tests

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/lib/integrationtest"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"
	"github.com/testcontainers/testcontainers-go"

	"auditor/internal/auditor/config"
)

const moduleID = "auditor"

var testEnv *environment

type environment struct {
	conf      config.Env
	connector mysql.Connector
	*integrationtest.Environment
}

func TestMain(m *testing.M) {
	testEnv = &environment{Environment: integrationtest.New()}
	os.Exit(integrationtest.Run(m, testEnv.Environment, testEnv.prepare))
}

func (e *environment) prepare(ctx context.Context) error {
	err := envconfig.Process("integration_defaults", &e.conf)
	if err != nil {
		return err
	}

	if err := e.StartNetwork(ctx); err != nil {
		return err
	}
	dsn, err := e.StartMySQL(ctx, moduleID, "mysql", "integration:password@local")
	if err != nil {
		return err
	}
	databaseConfig, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return err
	}
	e.conf.MySQL.Address = databaseConfig.Addr
	e.conf.MySQL.Database = databaseConfig.DBName
	e.conf.MySQL.User = databaseConfig.User
	e.conf.MySQL.Password = databaseConfig.Passwd

	e.connector = mysql.NewConnector()
	err = e.connector.Open(dsn, mysql.Config{
		MaxConnections: 10,
	})
	if err != nil {
		return err
	}
	e.AddCloser(e.connector.Close)

	e.conf.App.GraceTimeout = time.Second

	service, err := e.startApplication(ctx, "service")
	if err != nil {
		return err
	}
	e.conf.HTTP.Address, err = integrationtest.Address(ctx, service, "8082/tcp")
	if err != nil {
		return err
	}
	e.conf.GRPC.Address, err = integrationtest.Address(ctx, service, "8081/tcp")
	return err
}

func (e *environment) startApplication(ctx context.Context, command string) (testcontainers.Container, error) {
	prefix := strings.ToUpper(moduleID) + "_"
	return e.StartService(ctx, "distributed-programming-auditor:integration", command, map[string]string{
		prefix + "MYSQL_ADDRESS":     "mysql:3306",
		prefix + "MYSQL_DATABASE":    e.conf.MySQL.Database,
		prefix + "MYSQL_USER":        e.conf.MySQL.User,
		prefix + "MYSQL_PASSWORD":    e.conf.MySQL.Password,
		prefix + "HTTP_ADDRESS":      ":8082",
		prefix + "GRPC_ADDRESS":      ":8081",
		prefix + "APP_GRACE_TIMEOUT": e.conf.App.GraceTimeout.String(),
	})
}
