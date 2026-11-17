package integration_tests

import (
	"net/http"
	"testing"
	"time"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"

	api "planner/api/server/publicapi"
)

func TestProjectTaskLifecycle(t *testing.T) {
	client, err := api.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, err)
	user, workspace := uuid.New(), uuid.New()
	project := createTestProject(t, client, user, workspace)
	task := createTestTask(t, client, user, workspace, project.Id)
	verifyTaskUpdateConflict(t, client, user, workspace, task)
}

func createTestProject(t *testing.T, client *api.ClientWithResponses, user, workspace uuid.UUID) api.Project {
	t.Helper()
	response, err := client.CreateProjectWithResponse(t.Context(),
		&api.CreateProjectParams{XAuthUser: openapitypes.UUID(user), XWorkspaceID: openapitypes.UUID(workspace)},
		api.CreateProjectJSONRequestBody{Name: "Запуск"})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	return *response.JSON201
}

func createTestTask(t *testing.T, client *api.ClientWithResponses, user, workspace uuid.UUID, projectID openapitypes.UUID) api.Task {
	t.Helper()
	due := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	input := api.CreateTaskJSONRequestBody{
		Title: "Подготовить релиз", Description: "", Status: api.TaskInputStatusTodo,
		Priority: api.TaskInputPriorityHigh, DueAt: &due, SyncToCalendar: false,
	}
	response, err := client.CreateTaskWithResponse(t.Context(), projectID,
		&api.CreateTaskParams{XAuthUser: openapitypes.UUID(user), XWorkspaceID: openapitypes.UUID(workspace)}, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	return *response.JSON201
}

func verifyTaskUpdateConflict(t *testing.T, client *api.ClientWithResponses, user, workspace uuid.UUID, task api.Task) {
	t.Helper()
	input := api.UpdateTaskJSONRequestBody{
		Title: task.Title, Description: "Готово", Status: api.TaskUpdateStatusDone,
		Priority: api.TaskUpdatePriorityHigh, DueAt: task.DueAt, SyncToCalendar: false, Version: task.Version,
	}
	params := &api.UpdateTaskParams{XAuthUser: openapitypes.UUID(user), XWorkspaceID: openapitypes.UUID(workspace)}
	updated, err := client.UpdateTaskWithResponse(t.Context(), task.Id, params, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updated.StatusCode())
	require.EqualValues(t, 2, updated.JSON200.Version)
	conflict, err := client.UpdateTaskWithResponse(t.Context(), task.Id, params, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, conflict.StatusCode())
}
