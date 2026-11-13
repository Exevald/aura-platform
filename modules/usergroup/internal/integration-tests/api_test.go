package integration_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceScopedGroupLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	ownerID, workspaceID := uuid.New(), uuid.New()
	request := func(method, path, body string, workspace uuid.UUID) *http.Response {
		req, err := http.NewRequestWithContext(ctx, method, "http://"+testEnv.conf.HTTP.Address+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("X-Auth-User", ownerID.String())
		req.Header.Set("X-Workspace-ID", workspace.String())
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		response, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return response
	}

	created := request(http.MethodPost, "/api/usergroups/v1/groups", `{"name":"Engineering"}`, workspaceID)
	defer created.Body.Close()
	require.Equal(t, http.StatusCreated, created.StatusCode)
	var group struct {
		ID          uuid.UUID `json:"id"`
		WorkspaceID uuid.UUID `json:"workspaceID"`
		Name        string    `json:"name"`
	}
	require.NoError(t, json.NewDecoder(created.Body).Decode(&group))
	require.NotEqual(t, uuid.Nil, group.ID)
	require.Equal(t, workspaceID, group.WorkspaceID)
	require.Equal(t, "Engineering", group.Name)

	listed := request(http.MethodGet, "/api/usergroups/v1/groups", "", workspaceID)
	defer listed.Body.Close()
	require.Equal(t, http.StatusOK, listed.StatusCode)
	var groups []struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(listed.Body).Decode(&groups))
	require.Len(t, groups, 1)
	require.Equal(t, group.ID, groups[0].ID)

	otherWorkspace := request(http.MethodGet, fmt.Sprintf("/api/usergroups/v1/groups/%s", group.ID), "", uuid.New())
	defer otherWorkspace.Body.Close()
	require.Equal(t, http.StatusNotFound, otherWorkspace.StatusCode)
	deleted := request(http.MethodDelete, fmt.Sprintf("/api/usergroups/v1/groups/%s", group.ID), "", workspaceID)
	defer deleted.Body.Close()
	require.Equal(t, http.StatusNoContent, deleted.StatusCode)
	missing := request(http.MethodGet, fmt.Sprintf("/api/usergroups/v1/groups/%s", group.ID), "", workspaceID)
	defer missing.Body.Close()
	require.Equal(t, http.StatusNotFound, missing.StatusCode)
}
