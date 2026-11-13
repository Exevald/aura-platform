package public

import (
	api "calendar/api/server/publicapi"
	"calendar/internal/calendar/app"
	"calendar/internal/calendar/domain"
	"context"
	"errors"
	"time"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type handler struct{ svc *app.Service }

func NewHandler(s *app.Service) api.StrictServerInterface { return &handler{s} }

func (h *handler) ListCalendars(ctx context.Context, r api.ListCalendarsRequestObject) (api.ListCalendarsResponseObject, error) {
	v, e := h.svc.ListCalendars(ctx, uuid.UUID(r.Params.XWorkspaceID))
	if e != nil {
		return api.ListCalendars503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(api.ListCalendars200JSONResponse, 0, len(v))
	for _, x := range v {
		out = append(out, calendarAPI(x))
	}
	return out, nil
}

func (h *handler) CreateCalendar(ctx context.Context, r api.CreateCalendarRequestObject) (api.CreateCalendarResponseObject, error) {
	v, e := h.svc.CreateCalendar(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), r.Body.Name, r.Body.Timezone)
	if errors.Is(e, domain.ErrInvalid) {
		return api.CreateCalendar400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if e != nil {
		return api.CreateCalendar503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.CreateCalendar201JSONResponse(calendarAPI(v)), nil
}

func (h *handler) GetCalendar(ctx context.Context, r api.GetCalendarRequestObject) (api.GetCalendarResponseObject, error) {
	v, e := h.svc.GetCalendar(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.CalendarID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.GetCalendar404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.GetCalendar503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.GetCalendar200JSONResponse(calendarAPI(v)), nil
}

func (h *handler) UpdateCalendar(ctx context.Context, r api.UpdateCalendarRequestObject) (api.UpdateCalendarResponseObject, error) {
	v, e := h.svc.UpdateCalendar(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.CalendarID), r.Body.Name, r.Body.Timezone)
	if errors.Is(e, domain.ErrInvalid) {
		return api.UpdateCalendar400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if errors.Is(e, domain.ErrNotFound) {
		return api.UpdateCalendar404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.UpdateCalendar503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.UpdateCalendar200JSONResponse(calendarAPI(v)), nil
}

func (h *handler) DeleteCalendar(ctx context.Context, r api.DeleteCalendarRequestObject) (api.DeleteCalendarResponseObject, error) {
	e := h.svc.DeleteCalendar(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.CalendarID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.DeleteCalendar404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.DeleteCalendar503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.DeleteCalendar204Response{}, nil
}

func (h *handler) ListEvents(ctx context.Context, r api.ListEventsRequestObject) (api.ListEventsResponseObject, error) {
	var from, to zeroTime
	if r.Params.From != nil {
		from.t = *r.Params.From
	}
	if r.Params.To != nil {
		to.t = *r.Params.To
	}
	v, e := h.svc.ListEvents(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.CalendarID), from.t, to.t)
	if errors.Is(e, domain.ErrNotFound) {
		return api.ListEvents404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.ListEvents503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(api.ListEvents200JSONResponse, 0, len(v))
	for _, x := range v {
		out = append(out, eventAPI(x))
	}
	return out, nil
}

func (h *handler) CreateEvent(ctx context.Context, r api.CreateEventRequestObject) (api.CreateEventResponseObject, error) {
	v, e := h.svc.CreateEvent(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.CalendarID), eventInput(*r.Body, 0))
	if errors.Is(e, domain.ErrInvalid) {
		return api.CreateEvent400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if errors.Is(e, domain.ErrNotFound) {
		return api.CreateEvent404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.CreateEvent503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.CreateEvent201JSONResponse(eventAPI(v)), nil
}

func (h *handler) UpdateEvent(ctx context.Context, r api.UpdateEventRequestObject) (api.UpdateEventResponseObject, error) {
	v, e := h.svc.UpdateEvent(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.EventID), eventUpdate(*r.Body))
	if errors.Is(e, domain.ErrInvalid) {
		return api.UpdateEvent400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if errors.Is(e, domain.ErrNotFound) {
		return api.UpdateEvent404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if errors.Is(e, domain.ErrConflict) {
		return api.UpdateEvent409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return api.UpdateEvent503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.UpdateEvent200JSONResponse(eventAPI(v)), nil
}

func (h *handler) DeleteEvent(ctx context.Context, r api.DeleteEventRequestObject) (api.DeleteEventResponseObject, error) {
	e := h.svc.DeleteEvent(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.EventID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.DeleteEvent404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.DeleteEvent503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.DeleteEvent204Response{}, nil
}

type zeroTime struct{ t time.Time }

func calendarAPI(v domain.Calendar) api.Calendar {
	return api.Calendar{Id: openapi_types.UUID(v.ID), WorkspaceID: openapi_types.UUID(v.WorkspaceID), OwnerUserID: openapi_types.UUID(v.OwnerUserID), Name: v.Name, Timezone: v.Timezone, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func eventInput(v api.EventInput, version int64) domain.Event {
	out := domain.Event{Title: v.Title, Description: v.Description, StartsAt: v.StartsAt, EndsAt: v.EndsAt, Timezone: v.Timezone, Version: version}
	if v.Rrule != nil {
		out.RRule = *v.Rrule
	}
	for _, a := range v.Attendees {
		x := domain.Attendee{Email: string(a.Email), Status: string(a.Status)}
		if a.UserID != nil {
			x.UserID = uuid.UUID(*a.UserID)
		}
		out.Attendees = append(out.Attendees, x)
	}
	for _, m := range v.Reminders {
		out.Reminders = append(out.Reminders, domain.Reminder{MinutesBefore: m.MinutesBefore, InApp: m.InApp, Email: m.Email})
	}
	return out
}

func eventUpdate(v api.EventUpdate) domain.Event {
	return eventInput(api.EventInput{Title: v.Title, Description: v.Description, StartsAt: v.StartsAt, EndsAt: v.EndsAt, Timezone: v.Timezone, Rrule: v.Rrule, Attendees: v.Attendees, Reminders: v.Reminders}, v.Version)
}

func eventAPI(v domain.Event) api.Event {
	out := api.Event{Id: openapi_types.UUID(v.ID), CalendarID: openapi_types.UUID(v.CalendarID), Title: v.Title, Description: v.Description, StartsAt: v.StartsAt, EndsAt: v.EndsAt, Timezone: v.Timezone, Version: v.Version, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Attendees: make([]api.Attendee, 0, len(v.Attendees)), Reminders: make([]api.Reminder, 0, len(v.Reminders))}
	if v.RRule != "" {
		out.Rrule = &v.RRule
	}
	for _, a := range v.Attendees {
		x := api.Attendee{Email: openapi_types.Email(a.Email), Status: api.AttendeeStatus(a.Status)}
		if a.UserID != uuid.Nil() {
			id := openapi_types.UUID(a.UserID)
			x.UserID = &id
		}
		out.Attendees = append(out.Attendees, x)
	}
	for _, m := range v.Reminders {
		out.Reminders = append(out.Reminders, api.Reminder{MinutesBefore: m.MinutesBefore, InApp: m.InApp, Email: m.Email})
	}
	return out
}

func bad(e error) api.BadRequestJSONResponse { return api.BadRequestJSONResponse{Message: e.Error()} }

func notFound() api.NotFoundJSONResponse {
	return api.NotFoundJSONResponse{Message: "calendar resource not found"}
}

func unavailable() api.UnavailableJSONResponse {
	return api.UnavailableJSONResponse{Message: "calendar unavailable"}
}
