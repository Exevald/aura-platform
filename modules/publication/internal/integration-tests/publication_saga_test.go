package integration_tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	editorapi "editor/api/server/internalapi"
)

type releaseResponse struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	ShortPath      string `json:"shortPath"`
	DeliveryStatus string `json:"deliveryStatus"`
}

const documentIDField = "documentID"
const expectedVersionField = "expectedVersion"

func TestPublishSnapshotAndUnpublish(t *testing.T) {
	owner, member, workspace, document, file := uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()
	content := "## Первая версия\n\n![логотип](/api/editor/v1/documents/" + document + "/attachments/" + file + "/content)"
	testEnv.flowFakes.reset(&editorapi.PublicationSnapshot{DocumentId: document, Title: "Проверка саги", Content: content, Version: 1, Attachments: []*editorapi.PublicationAttachment{{FileId: file, FileName: "logo.png", ContentType: "image/png", SizeBytes: 7}}}, map[string][]byte{file: []byte("PNGDATA")}, []string{owner, member})
	key := uuid.New().String()
	resp, created := releaseRequest(t, http.MethodPost, "/api/publication/v1/releases", owner, workspace, key, map[string]any{documentIDField: document, expectedVersionField: 1})
	require.Equal(t, 202, resp)
	require.NotEmpty(t, created.ID)
	active := awaitRelease(t, owner, workspace, created.ID, "active")
	require.Equal(t, "sent", active.DeliveryStatus)
	require.Equal(t, "/s/Publish123", active.ShortPath)
	code, page := publicGet(t, "/p/"+created.ID)
	require.Equal(t, 200, code)
	require.Contains(t, string(page), "Первая версия")
	require.Contains(t, string(page), "/p/"+created.ID+"/assets/"+file)
	code, asset := publicGet(t, "/p/"+created.ID+"/assets/"+file)
	require.Equal(t, 200, code)
	require.Equal(t, []byte("PNGDATA"), asset)
	testEnv.flowFakes.mu.Lock()
	delete(testEnv.flowFakes.files, file)
	testEnv.flowFakes.mu.Unlock()
	code, asset = publicGet(t, "/p/"+created.ID+"/assets/"+file)
	require.Equal(t, 200, code)
	require.Equal(t, []byte("PNGDATA"), asset)

	// The same key must neither start another saga nor send duplicate notifications.
	resp, again := releaseRequest(t, http.MethodPost, "/api/publication/v1/releases", owner, workspace, key, map[string]any{documentIDField: document, expectedVersionField: 1})
	require.Equal(t, 202, resp)
	require.Equal(t, created.ID, again.ID)
	resp, _ = releaseRequest(t, http.MethodPost, "/api/publication/v1/releases", owner, workspace, key, map[string]any{documentIDField: uuid.New().String(), expectedVersionField: 1})
	require.Equal(t, 409, resp)
	testEnv.flowFakes.mu.Lock()
	links := len(testEnv.flowFakes.links)
	ownerNotifications := testEnv.flowFakes.notifications["publication:"+created.ID+":"+owner]
	memberNotifications := testEnv.flowFakes.notifications["publication:"+created.ID+":"+member]
	testEnv.flowFakes.mu.Unlock()
	require.Equal(t, 1, links)
	require.Equal(t, 1, ownerNotifications)
	require.Equal(t, 1, memberNotifications)

	resp, _ = releaseRequest(t, http.MethodDelete, "/api/publication/v1/releases/"+created.ID, owner, workspace, "", nil)
	require.Equal(t, 204, resp)
	awaitRelease(t, owner, workspace, created.ID, "closed")
	code, _ = publicGet(t, "/p/"+created.ID)
	require.Equal(t, 404, code)
	code, _ = publicGet(t, "/p/"+created.ID+"/assets/"+file)
	require.Equal(t, 404, code)
	testEnv.flowFakes.mu.Lock()
	remainingLinks := len(testEnv.flowFakes.links)
	testEnv.flowFakes.mu.Unlock()
	require.Zero(t, remainingLinks)
	testEnv.files.mu.Lock()
	require.Empty(t, testEnv.files.data)
	testEnv.files.mu.Unlock()
}

func TestPublishCompensatesChangedDocument(t *testing.T) {
	owner, workspace, document, file := uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()
	testEnv.flowFakes.reset(&editorapi.PublicationSnapshot{DocumentId: document, Title: "Меняется", Content: "текст", Version: 1, Attachments: []*editorapi.PublicationAttachment{{FileId: file, FileName: "file.txt", ContentType: "text/plain", SizeBytes: 4}}}, map[string][]byte{file: []byte("data")}, []string{owner})
	testEnv.flowFakes.changeAfterFirst = true
	resp, created := releaseRequest(t, http.MethodPost, "/api/publication/v1/releases", owner, workspace, uuid.New().String(), map[string]any{documentIDField: document, expectedVersionField: 1})
	require.Equal(t, 202, resp)
	awaitRelease(t, owner, workspace, created.ID, "failed")
	code, _ := publicGet(t, "/p/"+created.ID)
	require.Equal(t, 404, code)
	testEnv.flowFakes.mu.Lock()
	remainingLinks := len(testEnv.flowFakes.links)
	testEnv.flowFakes.mu.Unlock()
	require.Zero(t, remainingLinks)
	var assets int
	require.NoError(t, testEnv.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM publication_attachment WHERE release_id=?", created.ID).Scan(&assets))
	require.Zero(t, assets)
}

func TestShortlinkFailureDeletesExportedAttachment(t *testing.T) {
	owner, workspace, document, file := uuid.New().String(), uuid.New().String(), uuid.New().String(), uuid.New().String()
	testEnv.flowFakes.reset(&editorapi.PublicationSnapshot{DocumentId: document, Title: "Shortlink fails", Content: "file", Version: 1, Attachments: []*editorapi.PublicationAttachment{{FileId: file, FileName: "file.txt", ContentType: "text/plain", SizeBytes: 4}}}, map[string][]byte{file: []byte("data")}, []string{owner})
	testEnv.flowFakes.failLink = true
	status, created := releaseRequest(t, http.MethodPost, "/api/publication/v1/releases", owner, workspace, uuid.New().String(), map[string]any{documentIDField: document, expectedVersionField: 1})
	require.Equal(t, 202, status)
	awaitRelease(t, owner, workspace, created.ID, "failed")
	code, _ := publicGet(t, "/p/"+created.ID)
	require.Equal(t, 404, code)
	var assets int
	require.NoError(t, testEnv.db.QueryRowContext(t.Context(), "SELECT COUNT(*) FROM publication_attachment WHERE release_id=?", created.ID).Scan(&assets))
	require.Zero(t, assets)
	testEnv.files.mu.Lock()
	require.Empty(t, testEnv.files.data)
	testEnv.files.mu.Unlock()
}

func TestNotificationFailureKeepsPageActive(t *testing.T) {
	owner, workspace, document := uuid.New().String(), uuid.New().String(), uuid.New().String()
	testEnv.flowFakes.reset(&editorapi.PublicationSnapshot{DocumentId: document, Title: "Уведомления", Content: "открыто", Version: 1}, nil, []string{owner})
	testEnv.flowFakes.failNotifications = true
	resp, created := releaseRequest(t, http.MethodPost, "/api/publication/v1/releases", owner, workspace, uuid.New().String(), map[string]any{documentIDField: document, expectedVersionField: 1})
	require.Equal(t, 202, resp)
	awaitRelease(t, owner, workspace, created.ID, "active")
	require.Eventually(t, func() bool {
		r, v := releaseRequest(t, http.MethodGet, "/api/publication/v1/releases/"+created.ID, owner, workspace, "", nil)
		return r == 200 && v.DeliveryStatus == "failed"
	}, 30*time.Second, 100*time.Millisecond)
	code, page := publicGet(t, "/p/"+created.ID)
	require.Equal(t, 200, code)
	require.Contains(t, string(page), "открыто")
}

func releaseRequest(t *testing.T, method, path, owner, workspace, key string, body any) (int, releaseResponse) {
	t.Helper()
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(t.Context(), method, "http://"+testEnv.conf.HTTP.Address+path, payload)
	require.NoError(t, err)
	req.Header.Set("X-Auth-User", owner)
	req.Header.Set("X-Workspace-ID", workspace)
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	var result releaseResponse
	if resp.Header.Get("Content-Type") == "application/json" {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	}
	require.NoError(t, resp.Body.Close())
	return resp.StatusCode, result
}

func awaitRelease(t *testing.T, owner, workspace, id, status string) releaseResponse {
	t.Helper()
	var got releaseResponse
	require.Eventually(t, func() bool {
		resp, value := releaseRequest(t, http.MethodGet, "/api/publication/v1/releases/"+id, owner, workspace, "", nil)
		got = value
		return resp == 200 && value.Status == status
	}, 30*time.Second, 100*time.Millisecond, "release did not reach %s; last state: %+v", status, got)
	return got
}

func publicGet(t *testing.T, path string) (status int, body []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+testEnv.conf.HTTP.Address+path, http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, data
}
