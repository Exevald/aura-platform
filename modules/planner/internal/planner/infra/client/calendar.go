package client

import (
	api "calendar/api/server/internalapi"
	"context"
	"planner/internal/planner/app"
	"planner/internal/planner/domain"
	"time"
	"uuid"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type calendar struct {
	client  api.CalendarServiceClient
	timeout time.Duration
}

func NewCalendar(c api.CalendarServiceClient, t time.Duration) app.CalendarAdapter {
	return &calendar{c, t}
}

func (c *calendar) UpsertTaskEvent(ctx context.Context, w, a uuid.UUID, v domain.Task) (uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	start, end := app.TaskEventTimes(v)
	eventID := v.CalendarEventID
	if eventID == uuid.Nil() {
		eventID = v.ID
	}
	res, e := c.client.UpsertTaskEvent(ctx, &api.UpsertTaskEventRequest{IdempotencyKey: "planner:" + v.ID.String(), EventId: eventID.String(), CalendarId: v.CalendarID.String(), WorkspaceId: w.String(), ActorUserId: a.String(), Title: v.Title, Description: v.Description, StartsAt: timestamppb.New(start), EndsAt: timestamppb.New(end), Timezone: v.Timezone})
	if e != nil {
		return uuid.Nil(), e
	}
	return uuid.Parse(res.EventId)
}

func (c *calendar) DeleteTaskEvent(ctx context.Context, w, a, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	_, e := c.client.DeleteTaskEvent(ctx, &api.DeleteTaskEventRequest{EventId: id.String(), WorkspaceId: w.String(), ActorUserId: a.String()})
	return e
}
