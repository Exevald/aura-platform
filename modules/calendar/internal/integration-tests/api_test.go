package integration_tests

import (
	"net/http"
	"testing"
	"time"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"

	api "calendar/api/server/publicapi"
)

func TestCalendarEventLifecycle(t *testing.T) {
	client, err := api.NewClientWithResponses("http://" + testEnv.conf.HTTP.Address)
	require.NoError(t, err)
	user, workspace := uuid.New(), uuid.New()
	calendar := createTestCalendar(t, client, user, workspace)
	event := createRecurringEvent(t, client, user, workspace, calendar.Id)
	verifyCalendarEventUpdateConflict(t, client, user, workspace, event)
}

func createTestCalendar(t *testing.T, client *api.ClientWithResponses, user, workspace uuid.UUID) api.Calendar {
	t.Helper()
	response, err := client.CreateCalendarWithResponse(t.Context(),
		&api.CreateCalendarParams{XAuthUser: openapitypes.UUID(user), XWorkspaceID: openapitypes.UUID(workspace)},
		api.CreateCalendarJSONRequestBody{Name: "Команда", Timezone: "Europe/Moscow"})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	return *response.JSON201
}

func createRecurringEvent(t *testing.T, client *api.ClientWithResponses, user, workspace uuid.UUID, calendarID openapitypes.UUID) api.Event {
	t.Helper()
	start := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	rule := "FREQ=WEEKLY;COUNT=4"
	input := api.CreateEventJSONRequestBody{
		Title: "Планирование", Description: "Синхронизация", StartsAt: start, EndsAt: start.Add(time.Hour),
		Timezone: "Europe/Moscow", Rrule: &rule,
		Attendees: []api.Attendee{{Email: "guest@example.com", Status: api.Pending}},
		Reminders: []api.Reminder{{MinutesBefore: 15, InApp: true, Email: true}},
	}
	response, err := client.CreateEventWithResponse(t.Context(), calendarID,
		&api.CreateEventParams{XAuthUser: openapitypes.UUID(user), XWorkspaceID: openapitypes.UUID(workspace)}, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, response.StatusCode())
	require.EqualValues(t, 1, response.JSON201.Version)
	return *response.JSON201
}

func verifyCalendarEventUpdateConflict(t *testing.T, client *api.ClientWithResponses, user, workspace uuid.UUID, event api.Event) {
	t.Helper()
	input := api.UpdateEventJSONRequestBody{
		Title: event.Title, Description: "Новая повестка", StartsAt: event.StartsAt, EndsAt: event.EndsAt,
		Timezone: event.Timezone, Rrule: event.Rrule, Attendees: event.Attendees, Reminders: event.Reminders,
		Version: event.Version,
	}
	params := &api.UpdateEventParams{XAuthUser: openapitypes.UUID(user), XWorkspaceID: openapitypes.UUID(workspace)}
	updated, err := client.UpdateEventWithResponse(t.Context(), event.Id, params, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, updated.StatusCode())
	require.EqualValues(t, 2, updated.JSON200.Version)
	conflict, err := client.UpdateEventWithResponse(t.Context(), event.Id, params, input)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, conflict.StatusCode())
}
