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

	group := createTestGroup(t, request, workspaceID)
	verifyGroupListing(t, request, workspaceID, group.ID)
	verifyGroupWorkspaceIsolation(t, request, group.ID)
	verifyGroupDeletion(t, request, workspaceID, group.ID)
}

type groupRequest func(method, path, body string, workspace uuid.UUID) *http.Response

type testGroup struct {
	ID          uuid.UUID `json:"id"`
	WorkspaceID uuid.UUID `json:"workspaceID"`
	Name        string    `json:"name"`
}

func createTestGroup(t *testing.T, request groupRequest, workspace uuid.UUID) testGroup {
	t.Helper()
	created := request(http.MethodPost, "/api/usergroups/v1/groups", `{"name":"Engineering"}`, workspace)
	defer created.Body.Close()
	require.Equal(t, http.StatusCreated, created.StatusCode)
	var group testGroup
	require.NoError(t, json.NewDecoder(created.Body).Decode(&group))
	require.NotEqual(t, uuid.Nil, group.ID)
	require.Equal(t, workspace, group.WorkspaceID)
	require.Equal(t, "Engineering", group.Name)
	return group
}

func verifyGroupListing(t *testing.T, request groupRequest, workspaceID, groupID uuid.UUID) {
	t.Helper()
	listed := request(http.MethodGet, "/api/usergroups/v1/groups", "", workspaceID)
	defer listed.Body.Close()
	require.Equal(t, http.StatusOK, listed.StatusCode)
	var groups []struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(listed.Body).Decode(&groups))
	require.Len(t, groups, 1)
	require.Equal(t, groupID, groups[0].ID)
}

func verifyGroupWorkspaceIsolation(t *testing.T, request groupRequest, groupID uuid.UUID) {
	t.Helper()
	otherWorkspace := request(http.MethodGet, fmt.Sprintf("/api/usergroups/v1/groups/%s", groupID), "", uuid.New())
	defer otherWorkspace.Body.Close()
	require.Equal(t, http.StatusNotFound, otherWorkspace.StatusCode)
}

func verifyGroupDeletion(t *testing.T, request groupRequest, workspaceID, groupID uuid.UUID) {
	t.Helper()
	deleted := request(http.MethodDelete, fmt.Sprintf("/api/usergroups/v1/groups/%s", groupID), "", workspaceID)
	defer deleted.Body.Close()
	require.Equal(t, http.StatusNoContent, deleted.StatusCode)
	missing := request(http.MethodGet, fmt.Sprintf("/api/usergroups/v1/groups/%s", groupID), "", workspaceID)
	defer missing.Body.Close()
	require.Equal(t, http.StatusNotFound, missing.StatusCode)
}
