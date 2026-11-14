package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/distributed-programming-2026/go-sdk/pkg/migrator"
	"github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type version1791225168 struct{ db mysql.ClientContext }

func Version1791225168(db mysql.ClientContext) migrator.Migration { return &version1791225168{db} }

func (*version1791225168) Version() int64 {
	return 1791225168
}

func (*version1791225168) Description() string {
	return "Create calendars and events"
}

func (v *version1791225168) Up(ctx context.Context) error {
	const (
		createCalendarQuery = `
			CREATE TABLE IF NOT EXISTS calendar(
			    id BINARY(16) NOT NULL PRIMARY KEY,
			    workspace_id BINARY(16) NOT NULL,
			    owner_user_id BINARY(16) NOT NULL,
			    name VARCHAR(120) NOT NULL,
			    timezone VARCHAR(80) NOT NULL,
			    created_at DATETIME(6) NOT NULL,
			    updated_at DATETIME(6) NOT NULL,
			    INDEX calendar_workspace_idx(workspace_id)
			)
		`
		createCalendarEventQuery = `
			CREATE TABLE IF NOT EXISTS calendar_event(
			    id BINARY(16) NOT NULL PRIMARY KEY,
			    calendar_id BINARY(16) NOT NULL,
			    title VARCHAR(255) NOT NULL,
			    description TEXT NOT NULL,
			    starts_at DATETIME(6) NOT NULL,
			    ends_at DATETIME(6) NOT NULL,
			    timezone VARCHAR(80) NOT NULL,
			    rrule VARCHAR(512) NULL,
			    version BIGINT NOT NULL,
			    created_at DATETIME(6) NOT NULL,
			    updated_at DATETIME(6) NOT NULL,
			    INDEX calendar_event_time_idx(calendar_id, starts_at),
			    CONSTRAINT calendar_event_calendar_fk FOREIGN KEY(calendar_id) REFERENCES calendar(id) ON DELETE CASCADE
			)
		`
		createCalendarEventAttendeeQuery = `
			CREATE TABLE IF NOT EXISTS calendar_event_attendee(
			    event_id BINARY(16) NOT NULL,
			    user_id BINARY(16) NULL,
			    email VARCHAR(320) NOT NULL,
			    status ENUM('pending', 'accepted', 'declined', 'tentative') NOT NULL,
			    PRIMARY KEY(event_id, email),
			    CONSTRAINT attendee_event_fk FOREIGN KEY(event_id) REFERENCES calendar_event(id) ON DELETE CASCADE
			)
		`
		createCalendarEventReminderQuery = `
			CREATE TABLE IF NOT EXISTS calendar_event_reminder(
			    event_id BINARY(16) NOT NULL,
			    minutes_before INT NOT NULL,
			    in_app BOOLEAN NOT NULL,
			    email BOOLEAN NOT NULL,
			    PRIMARY KEY(event_id, minutes_before),
			    CONSTRAINT reminder_event_fk FOREIGN KEY(event_id) REFERENCES calendar_event(id) ON DELETE CASCADE
			)
		`
		createCalendarReminderDeliveryQuery = `
			CREATE TABLE IF NOT EXISTS calendar_reminder_delivery(
			    id BINARY(16) NOT NULL PRIMARY KEY,
			    event_id BINARY(16) NOT NULL,
			    workspace_id BINARY(16) NOT NULL,
			    user_id BINARY(16) NULL,
			    recipient_email VARCHAR(320) NOT NULL,
			    title VARCHAR(255) NOT NULL,
			    body TEXT NOT NULL,
			    rrule VARCHAR(512) NULL,
			    in_app BOOLEAN NOT NULL,
			    email BOOLEAN NOT NULL,
			    occurrence_starts_at DATETIME(6) NOT NULL,
			    deliver_at DATETIME(6) NOT NULL,
			    delivered_at DATETIME(6) NULL,
			    INDEX reminder_due_idx(delivered_at, deliver_at),
			    CONSTRAINT reminder_delivery_event_fk FOREIGN KEY(event_id) REFERENCES calendar_event(id) ON DELETE CASCADE
			)
		`
	)
	statements := []string{createCalendarQuery, createCalendarEventQuery, createCalendarEventAttendeeQuery, createCalendarEventReminderQuery, createCalendarReminderDeliveryQuery}

	for i, statement := range statements {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if _, err := v.db.ExecContext(ctx, fmt.Sprintf("calendar.schema.%d", i), statement); err != nil {
			return err
		}
	}
	return nil
}
