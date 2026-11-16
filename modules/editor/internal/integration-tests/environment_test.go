package integration_tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/distributed-programming-2026/lib/integrationtest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

type editorEnvironment struct {
	db      *sql.DB
	baseURL string
}

func newEditorEnvironment(t *testing.T) editorEnvironment {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	t.Cleanup(cancel)
	env := integrationtest.New()
	t.Cleanup(func() { require.NoError(t, env.Close()) })
	require.NoError(t, env.StartNetwork(ctx))
	dsn, err := env.StartMySQL(ctx, "editor", "mysql", "integration-password")
	require.NoError(t, err)
	_, err = env.StartBroker(ctx, "integration", "integration-password")
	require.NoError(t, err)
	service, err := env.StartService(
		ctx,
		"distributed-programming-editor:integration",
		"service",
		map[string]string{
			"EDITOR_MYSQL_ADDRESS":             "mysql:3306",
			"EDITOR_MYSQL_DATABASE":            "editor_test",
			"EDITOR_MYSQL_USER":                "editor",
			"EDITOR_MYSQL_PASSWORD":            "integration-password",
			"EDITOR_AMQP_URL":                  "amqp://integration:integration-password@rabbitmq:5672/",
			"EDITOR_HTTP_ADDRESS":              ":8082",
			"EDITOR_GRPC_ADDRESS":              ":8081",
			"EDITOR_DEPENDENCIES_USER_ADDRESS": "user-service:8081",
			"EDITOR_APP_GRACE_TIMEOUT":         "1s",
			"EDITOR_RELAY_INTERVAL":            "50ms",
			"EDITOR_RELAY_PUBLISH_TIMEOUT":     "1s",
			"EDITOR_RECOVERY_INTERVAL":         "50ms",
			"EDITOR_RECOVERY_ATTEMPT_TIMEOUT":  "1s",
			"EDITOR_RECOVERY_MAX_ATTEMPTS":     "3",
		})
	require.NoError(t, err, env.Logs.String())
	baseURL, err := service.PortEndpoint(ctx, "8082/tcp", "http")
	require.NoError(t, err)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	return editorEnvironment{db: db, baseURL: baseURL}
}
