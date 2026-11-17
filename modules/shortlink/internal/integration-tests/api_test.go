package integration_tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	publicapi "shortlink/api/server/publicapi"
)

func TestShortLinkLifecycle(t *testing.T) {
	baseURL := "http://" + testEnv.conf.HTTP.Address
	owner := uuid.New().String()
	link := createShortLink(t, baseURL, owner)
	redirectClient := &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	verifyShortLinkRedirect(t, redirectClient, baseURL+link.ShortPath, http.StatusFound)
	deleteShortLink(t, baseURL, owner, link.Id.String())
	verifyShortLinkRedirect(t, redirectClient, baseURL+link.ShortPath, http.StatusNotFound)
}

func createShortLink(t *testing.T, baseURL, owner string) publicapi.Link {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, baseURL+"/api/shortlink/v1/links", bytes.NewBufferString(`{"targetURL":"https://example.com/long?q=1"}`))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Auth-User", owner)
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	var link publicapi.Link
	require.NoError(t, json.NewDecoder(response.Body).Decode(&link))
	require.Len(t, link.Code, 10)
	return link
}

func verifyShortLinkRedirect(t *testing.T, client *http.Client, url string, expectedStatus int) {
	t.Helper()
	redirect, err := client.Get(url)
	require.NoError(t, err)
	defer redirect.Body.Close()
	require.Equal(t, expectedStatus, redirect.StatusCode)
	if expectedStatus == http.StatusFound {
		require.Equal(t, "https://example.com/long?q=1", redirect.Header.Get("Location"))
	}
}

func deleteShortLink(t *testing.T, baseURL, owner, id string) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodDelete, baseURL+"/api/shortlink/v1/links/"+id, http.NoBody)
	require.NoError(t, err)
	request.Header.Set("X-Auth-User", owner)
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestRejectsInvalidTarget(t *testing.T) {
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "http://"+testEnv.conf.HTTP.Address+"/api/shortlink/v1/links", bytes.NewBufferString(`{"targetURL":"javascript:alert(1)"}`))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Auth-User", uuid.New().String())
	response, err := (&http.Client{Timeout: time.Second}).Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
}
