package integration_tests

import (
	"net/http"
	api "planner/api/server/publicapi"
	"testing"
	"time"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestProjectTaskLifecycle(t *testing.T) {
	client, e := api.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, e)
	user, workspace := uuid.New(), uuid.New()
	project, e := client.CreateProjectWithResponse(t.Context(), &api.CreateProjectParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, api.CreateProjectJSONRequestBody{Name: "Запуск"})
	require.NoError(t, e)
	require.Equal(t, http.StatusCreated, project.StatusCode())
	due := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	task, e := client.CreateTaskWithResponse(t.Context(), project.JSON201.Id, &api.CreateTaskParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, api.CreateTaskJSONRequestBody{Title: "Подготовить релиз", Description: "", Status: api.TaskInputStatusTodo, Priority: api.TaskInputPriorityHigh, DueAt: &due, SyncToCalendar: false})
	require.NoError(t, e)
	require.Equal(t, http.StatusCreated, task.StatusCode())
	update := api.UpdateTaskJSONRequestBody{Title: task.JSON201.Title, Description: "Готово", Status: api.TaskUpdateStatusDone, Priority: api.TaskUpdatePriorityHigh, DueAt: task.JSON201.DueAt, SyncToCalendar: false, Version: task.JSON201.Version}
	res, e := client.UpdateTaskWithResponse(t.Context(), task.JSON201.Id, &api.UpdateTaskParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, update)
	require.NoError(t, e)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.EqualValues(t, 2, res.JSON200.Version)
	conflict, e := client.UpdateTaskWithResponse(t.Context(), task.JSON201.Id, &api.UpdateTaskParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, update)
	require.NoError(t, e)
	require.Equal(t, http.StatusConflict, conflict.StatusCode())
}
