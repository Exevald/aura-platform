package integration_tests

import (
	api "calendar/api/server/publicapi"
	"net/http"
	"testing"
	"time"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
)

func TestCalendarEventLifecycle(t *testing.T) {
	client, e := api.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, e)
	user, workspace := uuid.New(), uuid.New()
	params := &api.CreateCalendarParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}
	created, e := client.CreateCalendarWithResponse(t.Context(), params, api.CreateCalendarJSONRequestBody{Name: "Команда", Timezone: "Europe/Moscow"})
	require.NoError(t, e)
	require.Equal(t, http.StatusCreated, created.StatusCode())
	cal := created.JSON201
	start := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	rule := "FREQ=WEEKLY;COUNT=4"
	event, e := client.CreateEventWithResponse(t.Context(), cal.Id, &api.CreateEventParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, api.CreateEventJSONRequestBody{Title: "Планирование", Description: "Синхронизация", StartsAt: start, EndsAt: start.Add(time.Hour), Timezone: "Europe/Moscow", Rrule: &rule, Attendees: []api.Attendee{{Email: "guest@example.com", Status: api.Pending}}, Reminders: []api.Reminder{{MinutesBefore: 15, InApp: true, Email: true}}})
	require.NoError(t, e)
	require.Equal(t, http.StatusCreated, event.StatusCode())
	require.EqualValues(t, 1, event.JSON201.Version)
	updated := api.UpdateEventJSONRequestBody{Title: event.JSON201.Title, Description: "Новая повестка", StartsAt: event.JSON201.StartsAt, EndsAt: event.JSON201.EndsAt, Timezone: event.JSON201.Timezone, Rrule: event.JSON201.Rrule, Attendees: event.JSON201.Attendees, Reminders: event.JSON201.Reminders, Version: event.JSON201.Version}
	res, e := client.UpdateEventWithResponse(t.Context(), event.JSON201.Id, &api.UpdateEventParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, updated)
	require.NoError(t, e)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.EqualValues(t, 2, res.JSON200.Version)
	conflict, e := client.UpdateEventWithResponse(t.Context(), event.JSON201.Id, &api.UpdateEventParams{XAuthUser: openapi_types.UUID(user), XWorkspaceID: openapi_types.UUID(workspace)}, updated)
	require.NoError(t, e)
	require.Equal(t, http.StatusConflict, conflict.StatusCode())
}
