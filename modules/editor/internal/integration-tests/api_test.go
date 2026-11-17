package integration_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestServiceStartsAndOwnsDocumentPermissions(t *testing.T) {
	env := newEditorEnvironment(t)
	db := env.db
	for _, table := range []string{"document", "document_permission", "document_attachment"} {
		var count int
		err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='editor_test' AND table_name=?`, table).Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 1, count, table)
	}

	baseURL := env.baseURL
	ownerID, workspaceID := uuid.New(), uuid.New()
	request := func(method, path, body, workspace string) *http.Response {
		req, requestErr := http.NewRequestWithContext(t.Context(), method, baseURL+path, strings.NewReader(body))
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

	documentID := createEditorDocument(t, request, workspaceID)
	verifyEditorDocumentListing(t, request, workspaceID, documentID)
	verifyEditorWorkspaceIsolation(t, request, documentID)
}

type editorRequest func(method, path, body, workspace string) *http.Response

func createEditorDocument(t *testing.T, request editorRequest, workspaceID uuid.UUID) uuid.UUID {
	t.Helper()
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
	return document.ID
}

func verifyEditorDocumentListing(t *testing.T, request editorRequest, workspaceID, documentID uuid.UUID) {
	t.Helper()
	listed := request(http.MethodGet, "/api/editor/v1/documents", "", workspaceID.String())
	defer listed.Body.Close()
	require.Equal(t, http.StatusOK, listed.StatusCode)
	var documents []struct {
		ID uuid.UUID `json:"id"`
	}
	require.NoError(t, json.NewDecoder(listed.Body).Decode(&documents))
	require.Len(t, documents, 1)
	require.Equal(t, documentID, documents[0].ID)
}

func verifyEditorWorkspaceIsolation(t *testing.T, request editorRequest, documentID uuid.UUID) {
	t.Helper()
	otherWorkspace := uuid.New()
	crossWorkspace := request(http.MethodGet, fmt.Sprintf("/api/editor/v1/documents/%s", documentID), "", otherWorkspace.String())
	defer crossWorkspace.Body.Close()
	require.Equal(t, http.StatusNotFound, crossWorkspace.StatusCode)
}
