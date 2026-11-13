package domain

import (
	events "calendar/api/events"
	"net/mail"
	"strings"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"
)

type Service struct {
	repo   Repository
	events event.Dispatcher
}

func NewService(r Repository, e event.Dispatcher) *Service { return &Service{r, e} }

func validTimezone(v string) bool {
	_, e := time.LoadLocation(v)
	return e == nil
}

func validCalendar(name, tz string) bool {
	return strings.TrimSpace(name) != "" && len([]rune(name)) <= 120 && validTimezone(tz)
}

func validEvent(v Event) bool {
	if strings.TrimSpace(v.Title) == "" || len([]rune(v.Title)) > 255 || !v.StartsAt.Before(v.EndsAt) || !validTimezone(v.Timezone) {
		return false
	}
	if v.RRule != "" && !strings.HasPrefix(v.RRule, "FREQ=DAILY") && !strings.HasPrefix(v.RRule, "FREQ=WEEKLY") && !strings.HasPrefix(v.RRule, "FREQ=MONTHLY") {
		return false
	}
	for _, a := range v.Attendees {
		if _, e := mail.ParseAddress(a.Email); e != nil {
			return false
		}
		if a.Status != "pending" && a.Status != "accepted" && a.Status != "declined" && a.Status != "tentative" {
			return false
		}
	}
	for _, r := range v.Reminders {
		if r.MinutesBefore < 0 || r.MinutesBefore > 525600 || (!r.InApp && !r.Email) {
			return false
		}
	}
	return true
}

func (s *Service) CreateCalendar(workspace, actor uuid.UUID, name, tz string) (Calendar, error) {
	if workspace == uuid.Nil() || actor == uuid.Nil() || !validCalendar(name, tz) {
		return Calendar{}, ErrInvalid
	}
	now := time.Now().UTC()
	v := Calendar{ID: uuid.New(), WorkspaceID: workspace, OwnerUserID: actor, Name: strings.TrimSpace(name), Timezone: tz, CreatedAt: now, UpdatedAt: now}
	if e := s.repo.StoreCalendar(v); e != nil {
		return Calendar{}, e
	}
	e := s.events.Dispatch(&events.CalendarChanged{CalendarId: v.ID.String(), WorkspaceId: workspace.String(), ActorUserId: actor.String(), Action: "created", OccurredAtUnix: now.Unix()})
	return v, e
}

func (s *Service) List(workspace uuid.UUID) ([]Calendar, error) {
	if workspace == uuid.Nil() {
		return nil, ErrInvalid
	}
	return s.repo.ListCalendars(workspace)
}

func (s *Service) Get(workspace, id uuid.UUID) (Calendar, error) {
	v, e := s.repo.FindCalendar(id)
	if e != nil || v.WorkspaceID != workspace {
		return Calendar{}, ErrNotFound
	}
	return v, nil
}

func (s *Service) UpdateCalendar(workspace, actor, id uuid.UUID, name, tz string) (Calendar, error) {
	v, e := s.Get(workspace, id)
	if e != nil {
		return Calendar{}, e
	}
	if !validCalendar(name, tz) {
		return Calendar{}, ErrInvalid
	}
	v.Name, v.Timezone, v.UpdatedAt = strings.TrimSpace(name), tz, time.Now().UTC()
	if e = s.repo.StoreCalendar(v); e != nil {
		return Calendar{}, e
	}
	e = s.events.Dispatch(&events.CalendarChanged{CalendarId: id.String(), WorkspaceId: workspace.String(), ActorUserId: actor.String(), Action: "updated", OccurredAtUnix: v.UpdatedAt.Unix()})
	return v, e
}

func (s *Service) DeleteCalendar(workspace, actor, id uuid.UUID) error {
	if _, e := s.Get(workspace, id); e != nil {
		return e
	}
	if e := s.repo.DeleteCalendar(workspace, id); e != nil {
		return e
	}
	now := time.Now().UTC()
	return s.events.Dispatch(&events.CalendarChanged{CalendarId: id.String(), WorkspaceId: workspace.String(), ActorUserId: actor.String(), Action: "deleted", OccurredAtUnix: now.Unix()})
}

func (s *Service) CreateEvent(workspace, actor, calendarID uuid.UUID, v Event) (Event, error) {
	if _, e := s.Get(workspace, calendarID); e != nil {
		return Event{}, e
	}
	v.ID, v.CalendarID = uuid.New(), calendarID
	return s.createEvent(workspace, actor, v)
}

func (s *Service) CreateEventWithID(workspace, actor, id, calendarID uuid.UUID, v Event) (Event, error) {
	if _, e := s.Get(workspace, calendarID); e != nil {
		return Event{}, e
	}
	v.ID, v.CalendarID = id, calendarID
	if existing, e := s.repo.FindEvent(id); e == nil {
		v.Version = existing.Version
		return s.UpdateEvent(workspace, actor, id, v)
	}
	return s.createEvent(workspace, actor, v)
}

func (s *Service) createEvent(workspace, actor uuid.UUID, v Event) (Event, error) {
	if v.ID == uuid.Nil() || !validEvent(v) {
		return Event{}, ErrInvalid
	}
	now := time.Now().UTC()
	v.Version, v.CreatedAt, v.UpdatedAt = 1, now, now
	stored, e := s.repo.StoreEvent(v, 0)
	if e != nil {
		return Event{}, e
	}
	if e = s.repo.ReplaceReminderDeliveries(workspace, stored); e != nil {
		return Event{}, e
	}
	e = s.events.Dispatch(&events.EventChanged{EventId: v.ID.String(), CalendarId: v.CalendarID.String(), WorkspaceId: workspace.String(), ActorUserId: actor.String(), Action: "created", OccurredAtUnix: now.Unix()})
	return stored, e
}

func (s *Service) ListEvents(workspace, calendarID uuid.UUID, from, to time.Time) ([]Event, error) {
	if _, e := s.Get(workspace, calendarID); e != nil {
		return nil, e
	}
	return s.repo.ListEvents(calendarID, from, to)
}

func (s *Service) UpdateEvent(workspace, actor, id uuid.UUID, input Event) (Event, error) {
	old, e := s.repo.FindEvent(id)
	if e != nil {
		return Event{}, e
	}
	cal, e := s.Get(workspace, old.CalendarID)
	if e != nil {
		return Event{}, e
	}
	_ = cal
	if input.Version != old.Version || !validEvent(input) {
		if input.Version != old.Version {
			return Event{}, ErrConflict
		}
		return Event{}, ErrInvalid
	}
	input.ID, input.CalendarID, input.CreatedAt, input.UpdatedAt = id, old.CalendarID, old.CreatedAt, time.Now().UTC()
	stored, e := s.repo.StoreEvent(input, old.Version)
	if e != nil {
		return Event{}, e
	}
	if e = s.repo.ReplaceReminderDeliveries(workspace, stored); e != nil {
		return Event{}, e
	}
	e = s.events.Dispatch(&events.EventChanged{EventId: id.String(), CalendarId: old.CalendarID.String(), WorkspaceId: workspace.String(), ActorUserId: actor.String(), Action: "updated", OccurredAtUnix: input.UpdatedAt.Unix()})
	return stored, e
}

func (s *Service) DueReminders(now time.Time, limit int) ([]ReminderDelivery, error) {
	if limit <= 0 {
		return nil, ErrInvalid
	}
	return s.repo.DueReminderDeliveries(now, limit)
}

func (s *Service) CompleteReminder(v ReminderDelivery) error {
	next := nextOccurrence(v.OccurrenceStartsAt, v.RRule)
	if !next.IsZero() {
		return s.repo.CompleteReminderDelivery(v.ID, v.DeliverAt.Add(next.Sub(v.OccurrenceStartsAt)))
	}
	return s.repo.CompleteReminderDelivery(v.ID, time.Time{})
}

func nextOccurrence(at time.Time, rule string) time.Time {
	switch {
	case strings.HasPrefix(rule, "FREQ=DAILY"):
		return at.AddDate(0, 0, 1)
	case strings.HasPrefix(rule, "FREQ=WEEKLY"):
		return at.AddDate(0, 0, 7)
	case strings.HasPrefix(rule, "FREQ=MONTHLY"):
		return at.AddDate(0, 1, 0)
	default:
		return time.Time{}
	}
}

func (s *Service) DeleteEvent(workspace, actor, id uuid.UUID) error {
	v, e := s.repo.FindEvent(id)
	if e != nil {
		return e
	}
	if _, e = s.Get(workspace, v.CalendarID); e != nil {
		return e
	}
	if e = s.repo.DeleteEvent(id); e != nil {
		return e
	}
	now := time.Now().UTC()
	return s.events.Dispatch(&events.EventChanged{EventId: id.String(), CalendarId: v.CalendarID.String(), WorkspaceId: workspace.String(), ActorUserId: actor.String(), Action: "deleted", OccurredAtUnix: now.Unix()})
}
