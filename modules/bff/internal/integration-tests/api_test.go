package integration_tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/distributed-programming-2026/lib/integrationtest"
	"github.com/stretchr/testify/require"
)

func TestServiceHealthAndInvalidUserContext(t *testing.T) {
	env := integrationtest.New()
	t.Cleanup(func() { require.NoError(t, env.Close()) })
	require.NoError(t, env.StartNetwork(t.Context()))
	service, err := env.StartService(t.Context(), "distributed-programming-bff:integration", "service", map[string]string{
		"BFF_HTTP_ADDRESS":              ":8082",
		"BFF_GRPC_ADDRESS":              ":8081",
		"BFF_DEPENDENCIES_USER_ADDRESS": "missing-user:8081",
		"BFF_APP_GRACE_TIMEOUT":         "1s",
	})
	require.NoError(t, err, env.Logs.String())
	baseURL, err := service.PortEndpoint(t.Context(), "8082/tcp", "http")
	require.NoError(t, err)
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(baseURL + "/healthz")
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.NoError(t, response.Body.Close())
	response, err = client.Get(baseURL + "/api/bff/v1/me")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	require.NoError(t, response.Body.Close())
}
