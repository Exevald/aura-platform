package integration_tests

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"

	api "filestorage/api/server/publicapi"
)

func TestUploadDownloadDelete(t *testing.T) {
	user, workspace := uuid.New(), uuid.New()
	payload := []byte("attachment contents")
	file := uploadTestFile(t, user, workspace, payload)

	verifyDownloadedFile(t, user, workspace, file, payload)
	deleteTestFile(t, user, workspace, file)
}

func uploadTestFile(t *testing.T, user, workspace uuid.UUID, payload []byte) api.File {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, e := writer.CreateFormFile("file", "notes.txt")
	require.NoError(t, e)
	_, _ = part.Write(payload)
	require.NoError(t, writer.Close())

	request, e := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"http://"+testEnv.conf.HTTP.Address+"/api/files/v1/files",
		&body,
	)
	require.NoError(t, e)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-Auth-User", user.String())
	request.Header.Set("X-Workspace-ID", workspace.String())
	response, e := http.DefaultClient.Do(request)
	require.NoError(t, e)

	defer response.Body.Close()
	require.Equal(t, http.StatusCreated, response.StatusCode)
	var file api.File
	require.NoError(t, json.NewDecoder(response.Body).Decode(&file))
	return file
}

func verifyDownloadedFile(t *testing.T, user, workspace uuid.UUID, file api.File, payload []byte) {
	t.Helper()
	download, e := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"http://"+testEnv.conf.HTTP.Address+"/api/files/v1/files/"+file.Id.String()+"/content",
		http.NoBody,
	)
	require.NoError(t, e)
	download.Header.Set("X-Auth-User", user.String())
	download.Header.Set("X-Workspace-ID", workspace.String())
	downloadResponse, e := http.DefaultClient.Do(download)
	require.NoError(t, e)

	data, _ := io.ReadAll(downloadResponse.Body)
	_ = downloadResponse.Body.Close()
	require.Equal(t, payload, data)
}

func deleteTestFile(t *testing.T, user, workspace uuid.UUID, file api.File) {
	t.Helper()
	client, e := api.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, e)
	deleted, e := client.DeleteFileWithResponse(
		t.Context(),
		file.Id, &api.DeleteFileParams{
			XAuthUser:    openapitypes.UUID(user),
			XWorkspaceID: openapitypes.UUID(workspace),
		})
	require.NoError(t, e)
	require.Equal(t, http.StatusNoContent, deleted.StatusCode())
}
