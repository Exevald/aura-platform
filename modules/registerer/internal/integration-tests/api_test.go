package integration_tests

import (
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/distributed-programming-2026/lib/integrationtest"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

func TestServiceStartsMigratesAndRejectsInvalidRegistration(t *testing.T) {
	baseURL, db := startRegistererForTest(t)
	verifyRegistererMigration(t, db)
	verifyRegistererHealthAndValidation(t, baseURL)
}

func startRegistererForTest(t *testing.T) (string, *sql.DB) {
	t.Helper()
	env := integrationtest.New()
	t.Cleanup(func() { require.NoError(t, env.Close()) })
	require.NoError(t, env.StartNetwork(t.Context()))
	dsn, err := env.StartMySQL(t.Context(), "registerer", "mysql", "integration:password@local")
	require.NoError(t, err)
	_, err = env.StartBroker(t.Context(), "integration", "integration-password")
	require.NoError(t, err)
	service, err := env.StartService(t.Context(), "distributed-programming-registerer:integration", "service", map[string]string{
		"REGISTERER_MYSQL_ADDRESS":                  "mysql:3306",
		"REGISTERER_MYSQL_DATABASE":                 "registerer_test",
		"REGISTERER_MYSQL_USER":                     "registerer",
		"REGISTERER_MYSQL_PASSWORD":                 "integration:password@local",
		"REGISTERER_AMQP_URL":                       "amqp://integration:integration-password@rabbitmq:5672/",
		"REGISTERER_HTTP_ADDRESS":                   ":8082",
		"REGISTERER_GRPC_ADDRESS":                   ":8081",
		"REGISTERER_APP_GRACE_TIMEOUT":              "1s",
		"REGISTERER_RELAY_INTERVAL":                 "50ms",
		"REGISTERER_RELAY_PUBLISH_TIMEOUT":          "1s",
		"REGISTERER_RECOVERY_INTERVAL":              "50ms",
		"REGISTERER_DEPENDENCIES_AUTH_ADDRESS":      "missing-auth:8081",
		"REGISTERER_DEPENDENCIES_USER_ADDRESS":      "missing-user:8081",
		"REGISTERER_DEPENDENCIES_SESSION_ADDRESS":   "missing-session:8081",
		"REGISTERER_DEPENDENCIES_WORKSPACE_ADDRESS": "missing-workspace:8081",
		"REGISTERER_DEPENDENCIES_ACCESS_ADDRESS":    "missing-access:8081",
		"REGISTERER_COOKIE_SECURE":                  "false",
	})
	require.NoError(t, err, env.Logs.String())
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	baseURL, err := service.PortEndpoint(t.Context(), "8082/tcp", "http")
	require.NoError(t, err)
	return baseURL, db
}

func verifyRegistererMigration(t *testing.T, db *sql.DB) {
	t.Helper()
	var table string
	require.NoError(t, db.QueryRowContext(t.Context(), `SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`, "registerer_test", "registration_transaction").Scan(&table))
	require.Equal(t, "registration_transaction", table)
}

func verifyRegistererHealthAndValidation(t *testing.T, baseURL string) {
	t.Helper()
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(baseURL + "/healthz")
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.NoError(t, response.Body.Close())
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, baseURL+"/api/public/v1/registerer/register", strings.NewReader("{"))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	response, err = client.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	require.NoError(t, response.Body.Close())
}
