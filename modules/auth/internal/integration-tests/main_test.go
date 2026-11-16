package integration_tests

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
	"github.com/distributed-programming-2026/lib/integrationtest"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"auth/internal/auth/config"
)

const moduleID = "auth"

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

	e.conf.AMQP.URL, err = e.StartBroker(ctx, "integration", "integration-password")
	if err != nil {
		return err
	}
	if err = e.startDependencyDatabase(ctx, "session", "session-mysql"); err != nil {
		return err
	}
	if _, err = e.startDependencyApplication(ctx, "session", "session-mysql"); err != nil {
		return err
	}

	e.connector = mysql.NewConnector()
	err = e.connector.Open(dsn, mysql.Config{
		MaxConnections: 10,
	})
	if err != nil {
		return err
	}
	e.AddCloser(e.connector.Close)

	e.conf.App.GraceTimeout = time.Second
	e.conf.Relay.Interval = 50 * time.Millisecond
	e.conf.Relay.PublishTimeout = time.Second

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

func (e *environment) startDependencyDatabase(ctx context.Context, module, alias string) error {
	_, err := e.StartMySQL(ctx, module, alias, "integration:password@local")
	return err
}

func (e *environment) startDependencyApplication(ctx context.Context, module, mysqlAlias string) (testcontainers.Container, error) {
	prefix := strings.ToUpper(module) + "_"
	return e.StartContainer(ctx, testcontainers.ContainerRequest{
		Image:        "distributed-programming-" + module + ":integration",
		Cmd:          []string{"service"},
		ExposedPorts: []string{"8081/tcp", "8082/tcp"},
		Networks:     []string{e.Network.Name},
		NetworkAliases: map[string][]string{
			e.Network.Name: {module},
		},
		Env: map[string]string{
			prefix + "MYSQL_ADDRESS":         mysqlAlias + ":3306",
			prefix + "MYSQL_DATABASE":        module + "_test",
			prefix + "MYSQL_USER":            module,
			prefix + "MYSQL_PASSWORD":        "integration:password@local",
			prefix + "AMQP_URL":              "amqp://integration:integration-password@rabbitmq:5672/",
			prefix + "HTTP_ADDRESS":          ":8082",
			prefix + "GRPC_ADDRESS":          ":8081",
			prefix + "APP_GRACE_TIMEOUT":     e.conf.App.GraceTimeout.String(),
			prefix + "RELAY_INTERVAL":        e.conf.Relay.Interval.String(),
			prefix + "RELAY_PUBLISH_TIMEOUT": e.conf.Relay.PublishTimeout.String(),
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&e.Logs},
		},
		WaitingFor: wait.ForHTTP("/healthz").WithPort("8082/tcp").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusNoContent
		}).WithStartupTimeout(time.Minute),
	})
}

func (e *environment) startApplication(ctx context.Context, command string) (testcontainers.Container, error) {
	prefix := strings.ToUpper(moduleID) + "_"
	return e.StartService(ctx, "distributed-programming-auth:integration", command, map[string]string{
		prefix + "MYSQL_ADDRESS":                "mysql:3306",
		prefix + "MYSQL_DATABASE":               e.conf.MySQL.Database,
		prefix + "MYSQL_USER":                   e.conf.MySQL.User,
		prefix + "MYSQL_PASSWORD":               e.conf.MySQL.Password,
		prefix + "AMQP_URL":                     "amqp://integration:integration-password@rabbitmq:5672/",
		prefix + "HTTP_ADDRESS":                 ":8082",
		prefix + "GRPC_ADDRESS":                 ":8081",
		prefix + "APP_GRACE_TIMEOUT":            e.conf.App.GraceTimeout.String(),
		prefix + "RELAY_INTERVAL":               e.conf.Relay.Interval.String(),
		prefix + "RELAY_PUBLISH_TIMEOUT":        e.conf.Relay.PublishTimeout.String(),
		prefix + "DEPENDENCIES_SESSION_ADDRESS": "session:8081",
		prefix + "DEPENDENCIES_TIMEOUT":         "5s",
		prefix + "COOKIE_SECURE":                "false",
	})
}
