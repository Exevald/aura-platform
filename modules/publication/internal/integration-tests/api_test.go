package integration_tests

import (
	"net/http"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestReleaseRequestValidation(t *testing.T) {
	endpoint := "http://" + testEnv.conf.HTTP.Address + "/api/publication/v1/releases"
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, strings.NewReader(`{"documentID":"bad","expectedVersion":1}`))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusUnauthorized, response.StatusCode)

	request, err = http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, strings.NewReader(`{"documentID":"bad","expectedVersion":1}`))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Auth-User", uuid.New().String())
	request.Header.Set("X-Workspace-ID", uuid.New().String())
	request.Header.Set("Idempotency-Key", uuid.New().String())
	response, err = http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestPublicRoutesRejectInvalidIDs(t *testing.T) {
	for _, path := range []string{"/p/invalid", "/p/" + uuid.New().String() + "/assets/invalid"} {
		request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+testEnv.conf.HTTP.Address+path, http.NoBody)
		require.NoError(t, err)
		response, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		require.Equal(t, http.StatusNotFound, response.StatusCode)
	}
}
