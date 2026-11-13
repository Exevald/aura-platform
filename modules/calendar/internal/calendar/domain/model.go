package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalid  = errors.New("invalid calendar data")
	ErrNotFound = errors.New("calendar resource not found")
	ErrConflict = errors.New("calendar resource version conflict")
)

type Calendar struct {
	ID, WorkspaceID, OwnerUserID uuid.UUID
	Name, Timezone               string
	CreatedAt, UpdatedAt         time.Time
}

type Attendee struct {
	UserID        uuid.UUID
	Email, Status string
}

type Reminder struct {
	MinutesBefore int
	InApp, Email  bool
}

type Event struct {
	ID, CalendarID       uuid.UUID
	Title, Description   string
	StartsAt, EndsAt     time.Time
	Timezone, RRule      string
	Attendees            []Attendee
	Reminders            []Reminder
	Version              int64
	CreatedAt, UpdatedAt time.Time
}

type ReminderDelivery struct {
	ID, EventID, WorkspaceID, UserID   uuid.UUID
	RecipientEmail, Title, Body, RRule string
	InApp, Email                       bool
	OccurrenceStartsAt, DeliverAt      time.Time
}

type Repository interface {
	StoreCalendar(Calendar) error
	FindCalendar(uuid.UUID) (Calendar, error)
	ListCalendars(uuid.UUID) ([]Calendar, error)
	DeleteCalendar(uuid.UUID, uuid.UUID) error
	StoreEvent(Event, int64) (Event, error)
	FindEvent(uuid.UUID) (Event, error)
	ListEvents(uuid.UUID, time.Time, time.Time) ([]Event, error)
	DeleteEvent(uuid.UUID) error
	ReplaceReminderDeliveries(uuid.UUID, Event) error
	DueReminderDeliveries(time.Time, int) ([]ReminderDelivery, error)
	CompleteReminderDelivery(uuid.UUID, time.Time) error
}
