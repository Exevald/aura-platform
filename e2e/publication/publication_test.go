package publication

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type apiClient struct {
	base string
	http *http.Client
}

func newClient(base string) *apiClient {
	jar, _ := cookiejar.New(nil)
	return &apiClient{strings.TrimRight(base, "/"), &http.Client{Jar: jar, Timeout: 10 * time.Second}}
}
func (c *apiClient) call(ctx context.Context, method, path, workspace, key string, body any) (status int, data []byte, headers http.Header, callErr error) {
	var input io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, nil, err
		}
		input = bytes.NewReader(raw)
	}
	// #nosec G704 -- this test targets the explicitly configured local gateway.
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, input)
	if err != nil {
		return 0, nil, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if workspace != "" {
		req.Header.Set("X-Workspace-ID", workspace)
	}
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	// #nosec G704 -- the request URL comes from the E2E gateway setting.
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	return resp.StatusCode, data, resp.Header, err
}
func callJSON[T any](t *testing.T, c *apiClient, method, path, workspace, key string, body any, want int) T {
	t.Helper()
	status, data, _, err := c.call(t.Context(), method, path, workspace, key, body)
	require.NoError(t, err)
	require.Equal(t, want, status, "%s %s: %s", method, path, data)
	var result T
	if len(data) > 0 {
		require.NoError(t, json.Unmarshal(data, &result))
	}
	return result
}
func await[T any](t *testing.T, timeout time.Duration, check func() (T, bool)) T {
	t.Helper()
	var result T
	require.Eventually(t, func() bool { var ok bool; result, ok = check(); return ok }, timeout, 200*time.Millisecond)
	return result
}

type registered struct {
	UserID string `json:"userID"`
}
type workspace struct {
	ID string `json:"id"`
}
type group struct {
	ID string `json:"id"`
}
type document struct {
	ID      string `json:"id"`
	Version int64  `json:"version"`
}
type release struct {
	ID             string `json:"id"`
	Status         string `json:"status"`
	ShortPath      string `json:"shortPath"`
	DeliveryStatus string `json:"deliveryStatus"`
}
type notification struct {
	Kind string `json:"kind"`
	Body string `json:"body"`
}

const documentIDField = "documentID"
const expectedVersionField = "expectedVersion"
const emailField = "email"

func TestPublicationLifecycleThroughGateway(t *testing.T) {
	base := os.Getenv("AURA_E2E_BASE_URL")
	if base == "" {
		t.Skip("set AURA_E2E_BASE_URL to a running full Aura stack")
	}
	_, err := url.ParseRequestURI(base)
	require.NoError(t, err)
	owner, member := newClient(base), newClient(base)
	unique := uuid.New().String()
	register := func(c *apiClient, name string) string {
		login := "publication-" + name + "-" + unique
		result := callJSON[registered](t, c, http.MethodPost, "/api/public/v1/registerer/register", "", "", map[string]any{
			"login": login, "password": "Test-password-123!", "displayName": name, "firstName": name, "lastName": "E2E", emailField: login + "@example.com", "rememberMe": false,
		}, 201)
		require.NotEmpty(t, result.UserID)
		return result.UserID
	}
	_, memberID := register(owner, "owner"), register(member, "member")
	w := callJSON[workspace](t, owner, http.MethodPost, "/api/workspaces/v1/workspaces", "", "", map[string]string{"name": "Publication E2E " + unique}, 201)
	require.NotEmpty(t, w.ID)
	callJSON[workspace](t, owner, http.MethodPut, "/api/workspaces/v1/workspaces/"+w.ID+"/members", "", "", map[string]string{"userID": memberID, "role": "viewer"}, 200)
	g := callJSON[group](t, owner, http.MethodPost, "/api/usergroups/v1/groups", w.ID, "", map[string]string{"name": "Publication access " + unique}, 201)
	ownerEmail := "publication-owner-" + unique + "@example.com"
	memberEmail := "publication-member-" + unique + "@example.com"
	callJSON[group](t, owner, http.MethodPost, "/api/usergroups/v1/groups/"+g.ID+"/members", w.ID, "", map[string]string{emailField: ownerEmail}, 200)
	callJSON[group](t, owner, http.MethodPost, "/api/usergroups/v1/groups/"+g.ID+"/members", w.ID, "", map[string]string{emailField: memberEmail}, 200)
	callJSON[map[string]any](t, owner, http.MethodPut, "/api/access/v1/groups/"+g.ID+"/modules", "", "", map[string]any{"moduleIDs": []string{
		"b8547c76-51fd-4f7c-9c68-8f30c84f61cb", // Editor
		"84c7d60e-6f3e-4f09-91c2-82774c12fc88", // Publication
	}}, 200)

	var doc document
	var lastDocumentStatus int
	var lastDocumentResponse []byte
	var lastDocumentErr error
	createdDocument := assert.Eventually(t, func() bool {
		status, data, _, err := owner.call(t.Context(), http.MethodPost, "/api/editor/v1/documents", w.ID, "", map[string]any{"title": "Первая версия", "content": "## Снимок один"})
		lastDocumentStatus, lastDocumentResponse, lastDocumentErr = status, data, err
		if err != nil || status != 201 {
			return false
		}
		if json.Unmarshal(data, &doc) != nil {
			return false
		}
		return true
	}, 30*time.Second, 200*time.Millisecond)
	if !createdDocument {
		t.Fatalf("document creation returned %d: %s (%v)", lastDocumentStatus, lastDocumentResponse, lastDocumentErr)
	}
	require.NotEmpty(t, doc.ID)
	key := uuid.New().String()
	r := callJSON[release](t, owner, http.MethodPost, "/api/publication/v1/releases", w.ID, key, map[string]any{documentIDField: doc.ID, expectedVersionField: doc.Version}, 202)
	require.NotEmpty(t, r.ID)
	getRelease := func(id string) release {
		return callJSON[release](t, owner, http.MethodGet, "/api/publication/v1/releases/"+id, w.ID, "", nil, 200)
	}
	r = await(t, 45*time.Second, func() (release, bool) {
		v := getRelease(r.ID)
		return v, v.Status == "active" && v.DeliveryStatus == "sent"
	})
	require.NotEmpty(t, r.ShortPath)
	await(t, 30*time.Second, func() (bool, bool) {
		status, _, _, err := member.call(t.Context(), http.MethodGet, "/api/publication/v1/releases", w.ID, "", nil)
		return true, err == nil && status == 200
	})
	memberReadStatus, _, _, err := member.call(t.Context(), http.MethodGet, "/api/publication/v1/releases/"+r.ID, w.ID, "", nil)
	require.NoError(t, err)
	require.Equal(t, 404, memberReadStatus)
	guest := newClient(base)
	status, page, _, err := guest.call(t.Context(), http.MethodGet, "/p/"+r.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 200, status)
	require.Contains(t, string(page), "Снимок один")
	redirectClient := newClient(base)
	redirectClient.http.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	status, _, headers, err := redirectClient.call(t.Context(), http.MethodGet, r.ShortPath, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 302, status)
	require.Contains(t, headers.Get("Location"), "/p/"+r.ID)

	updated := callJSON[document](t, owner, http.MethodPut, "/api/editor/v1/documents/"+doc.ID, w.ID, "", map[string]any{"title": "Вторая версия", "content": "## Снимок два", "version": doc.Version}, 200)
	status, page, _, err = guest.call(t.Context(), http.MethodGet, "/p/"+r.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 200, status)
	require.Contains(t, string(page), "Снимок один")
	require.NotContains(t, string(page), "Снимок два")
	memberStatus, memberData, _, err := member.call(t.Context(), http.MethodPost, "/api/publication/v1/releases", w.ID, uuid.New().String(), map[string]any{documentIDField: doc.ID, expectedVersionField: updated.Version})
	require.NoError(t, err)
	require.Equal(t, 202, memberStatus, string(memberData))
	var denied release
	require.NoError(t, json.Unmarshal(memberData, &denied))
	await(t, 30*time.Second, func() (bool, bool) {
		status, data, _, err := member.call(t.Context(), http.MethodGet, "/api/publication/v1/releases/"+denied.ID, w.ID, "", nil)
		if err != nil || status != 200 {
			return false, false
		}
		var current release
		return true, json.Unmarshal(data, &current) == nil && current.Status == "failed"
	})
	for _, c := range []*apiClient{owner, member} {
		await(t, 30*time.Second, func() (bool, bool) {
			status, data, _, err := c.call(t.Context(), http.MethodGet, "/api/notifications/v1/notifications", "", "", nil)
			if err != nil || status != 200 {
				return false, false
			}
			var values []notification
			if json.Unmarshal(data, &values) != nil {
				return false, false
			}
			for _, v := range values {
				if v.Kind == "publication" && strings.Contains(v.Body, r.ShortPath) {
					return true, true
				}
			}
			return false, false
		})
	}

	status, _, _, err = owner.call(t.Context(), http.MethodDelete, "/api/publication/v1/releases/"+r.ID, w.ID, "", nil)
	require.NoError(t, err)
	require.Equal(t, 204, status)
	await(t, 45*time.Second, func() (bool, bool) { return true, getRelease(r.ID).Status == "closed" })
	status, _, _, err = guest.call(t.Context(), http.MethodGet, "/p/"+r.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 404, status)
	status, _, _, err = redirectClient.call(t.Context(), http.MethodGet, r.ShortPath, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 404, status)

	second := callJSON[release](t, owner, http.MethodPost, "/api/publication/v1/releases", w.ID, uuid.New().String(), map[string]any{documentIDField: doc.ID, expectedVersionField: updated.Version}, 202)
	second = await(t, 45*time.Second, func() (release, bool) { v := getRelease(second.ID); return v, v.Status == "active" })
	require.NotEqual(t, r.ShortPath, second.ShortPath)
	status, _, _, err = owner.call(t.Context(), http.MethodDelete, "/api/editor/v1/documents/"+doc.ID, w.ID, "", nil)
	require.NoError(t, err)
	require.Equal(t, 204, status)
	await(t, 45*time.Second, func() (bool, bool) { return true, getRelease(second.ID).Status == "closed" })
	status, _, _, err = guest.call(t.Context(), http.MethodGet, "/p/"+second.ID, "", "", nil)
	require.NoError(t, err)
	require.Equal(t, 404, status)
}
