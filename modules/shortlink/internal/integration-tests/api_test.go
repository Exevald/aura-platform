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

	redirectClient := &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	redirect, err := redirectClient.Get(baseURL + link.ShortPath)
	require.NoError(t, err)
	defer redirect.Body.Close()
	require.Equal(t, http.StatusFound, redirect.StatusCode)
	require.Equal(t, "https://example.com/long?q=1", redirect.Header.Get("Location"))

	request, err = http.NewRequestWithContext(t.Context(), http.MethodDelete, baseURL+"/api/shortlink/v1/links/"+link.Id.String(), http.NoBody)
	require.NoError(t, err)
	request.Header.Set("X-Auth-User", owner)
	response, err = (&http.Client{Timeout: time.Second}).Do(request)
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusNoContent, response.StatusCode)

	redirect, err = redirectClient.Get(baseURL + link.ShortPath)
	require.NoError(t, err)
	defer redirect.Body.Close()
	require.Equal(t, http.StatusNotFound, redirect.StatusCode)
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
