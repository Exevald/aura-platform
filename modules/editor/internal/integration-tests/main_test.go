package integration_tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestServiceStartsAndOwnsDocumentPermissions(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	testNetwork, err := network.New(ctx)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testNetwork.Remove(context.Background())) })

	database, err := tcmysql.Run(ctx, "mysql:8.4",
		tcmysql.WithDatabase("editor_test"),
		tcmysql.WithUsername("editor"),
		tcmysql.WithPassword("integration-password"),
		network.WithNetwork([]string{"mysql"}, testNetwork),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(database)) })

	broker, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started:  true,
		Image:    "rabbitmq:4.1-alpine",
		Networks: []string{testNetwork.Name},
		NetworkAliases: map[string][]string{
			testNetwork.Name: {"rabbitmq"},
		},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": "integration",
			"RABBITMQ_DEFAULT_PASS": "integration-password",
		},
		WaitingFor: wait.ForLog("Server startup complete").WithStartupTimeout(time.Minute),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(broker)) })

	service, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started:      true,
		Image:        "distributed-programming-editor:integration",
		ExposedPorts: []string{"8082/tcp"},
		Networks:     []string{testNetwork.Name},
		Env: map[string]string{
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
		},
		WaitingFor: wait.ForHTTP("/healthz").WithPort("8082/tcp").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusNoContent
		}).WithStartupTimeout(time.Minute),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, testcontainers.TerminateContainer(service)) })

	dsn, err := database.ConnectionString(ctx, "parseTime=true")
	require.NoError(t, err)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	for _, table := range []string{"document", "document_permission", "document_attachment"} {
		var count int
		err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='editor_test' AND table_name=?`, table).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, table)
	}

	baseURL, err := service.PortEndpoint(ctx, "8082/tcp", "http")
	require.NoError(t, err)
	ownerID, workspaceID := uuid.New(), uuid.New()
	request := func(method, path, body, workspace string) *http.Response {
		req, requestErr := http.NewRequestWithContext(ctx, method, baseURL+path, strings.NewReader(body))
		require.NoError(t, requestErr)
		req.Header.Set("X-Auth-User", ownerID.String())
		req.Header.Set("X-Workspace-ID", workspace)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		response, requestErr := http.DefaultClient.Do(req)
		require.NoError(t, requestErr)
		return response
	}

	created := request(http.MethodPost, "/api/editor/v1/documents", `{"title":"Workspace notes","content":"First draft"}`, workspaceID.String())
	defer created.Body.Close()
	require.Equal(t, http.StatusCreated, created.StatusCode)
	var document struct {
		ID          uuid.UUID `json:"id"`
		WorkspaceID uuid.UUID `json:"workspaceID"`
		Title       string    `json:"title"`
		Version     int64     `json:"version"`
	}
	require.NoError(t, json.NewDecoder(created.Body).Decode(&document))
	require.NotEqual(t, uuid.Nil, document.ID)
	require.Equal(t, workspaceID, document.WorkspaceID)
	require.Equal(t, "Workspace notes", document.Title)
	require.EqualValues(t, 1, document.Version)

	listed := request(http.MethodGet, "/api/editor/v1/documents", "", workspaceID.String())
	defer listed.Body.Close()
	require.Equal(t, http.StatusOK, listed.StatusCode)
	var documents []struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(listed.Body).Decode(&documents))
	require.Len(t, documents, 1)
	require.Equal(t, document.ID, documents[0].ID)

	otherWorkspace := uuid.New()
	crossWorkspace := request(http.MethodGet, fmt.Sprintf("/api/editor/v1/documents/%s", document.ID), "", otherWorkspace.String())
	defer crossWorkspace.Body.Close()
	require.Equal(t, http.StatusNotFound, crossWorkspace.StatusCode)
}
