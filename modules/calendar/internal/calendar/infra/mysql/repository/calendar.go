package repository

import (
	"calendar/internal/calendar/domain"
	"context"
	"database/sql"
	"errors"
	"time"
	"uuid"

	sdkmysql "github.com/distributed-programming-2026/go-sdk/pkg/mysql"
)

type repo struct {
	ctx context.Context
	db  sdkmysql.ClientContext
}

func New(ctx context.Context, db sdkmysql.ClientContext) domain.Repository { return &repo{ctx, db} }

type calendarRow struct {
	ID          uuid.UUID `db:"id"`
	WorkspaceID uuid.UUID `db:"workspace_id"`
	OwnerUserID uuid.UUID `db:"owner_user_id"`
	Name        string    `db:"name"`
	Timezone    string    `db:"timezone"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type eventRow struct {
	ID          uuid.UUID      `db:"id"`
	CalendarID  uuid.UUID      `db:"calendar_id"`
	Title       string         `db:"title"`
	Description string         `db:"description"`
	StartsAt    time.Time      `db:"starts_at"`
	EndsAt      time.Time      `db:"ends_at"`
	Timezone    string         `db:"timezone"`
	RRule       sql.NullString `db:"rrule"`
	Version     int64          `db:"version"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (r *repo) StoreCalendar(v domain.Calendar) error {
	const sqlCalendarStoreQuery = `
		INSERT INTO calendar(
		    id,
		    workspace_id,
		    owner_user_id,
		    name,
		    timezone,
		    created_at,
		    updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		    name = VALUES (name),
		    timezone = VALUES (timezone),
		    updated_at = VALUES (updated_at)
	`

	_, e := r.db.ExecContext(r.ctx, "calendar.store", sqlCalendarStoreQuery, v.ID, v.WorkspaceID, v.OwnerUserID, v.Name, v.Timezone, v.CreatedAt, v.UpdatedAt)
	return e
}

func (r *repo) FindCalendar(id uuid.UUID) (domain.Calendar, error) {
	const sqlCalendarFindQuery = `
		SELECT
		    id,
		    workspace_id,
		    owner_user_id,
		    name,
		    timezone,
		    created_at,
		    updated_at
		FROM calendar
		WHERE id = ?
	`

	var v calendarRow
	e := r.db.GetContext(r.ctx, "calendar.find", &v, sqlCalendarFindQuery, id)
	if errors.Is(e, sql.ErrNoRows) {
		e = domain.ErrNotFound
	}
	return domain.Calendar(v), e
}

func (r *repo) ListCalendars(w uuid.UUID) ([]domain.Calendar, error) {
	const sqlCalendarListQuery = `
		SELECT
		    id,
		    workspace_id,
		    owner_user_id,
		    name,
		    timezone,
		    created_at,
		    updated_at
		FROM calendar
		WHERE workspace_id = ?
		ORDER BY name
	`

	var rows []calendarRow
	e := r.db.SelectContext(r.ctx, "calendar.list", &rows, sqlCalendarListQuery, w)
	out := make([]domain.Calendar, 0, len(rows))
	for _, v := range rows {
		out = append(out, domain.Calendar(v))
	}
	return out, e
}

func (r *repo) DeleteCalendar(w, id uuid.UUID) error {
	const sqlCalendarDeleteQuery = `
		DELETE FROM calendar
		WHERE id = ? AND workspace_id = ?
	`

	res, e := r.db.ExecContext(r.ctx, "calendar.delete", sqlCalendarDeleteQuery, id, w)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

func (r *repo) StoreEvent(v domain.Event, expected int64) (domain.Event, error) {
	const (
		sqlEventInsertQuery = `
			INSERT INTO calendar_event(
			    id,
			    calendar_id,
			    title,
			    description,
			    starts_at,
			    ends_at,
			    timezone,
			    rrule,
			    version,
			    created_at,
			    updated_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		sqlEventUpdateQuery = `
			UPDATE calendar_event
			SET
			    title = ?,
			    description = ?,
			    starts_at = ?,
			    ends_at = ?,
			    timezone = ?,
			    rrule = ?,
			    version = version+1,
			    updated_at = ?
			WHERE id = ? AND version = ?
		`
		sqlAttendeeReplaceDeleteQuery = `
			DELETE FROM calendar_event_attendee
			WHERE event_id = ?
		`
		sqlAttendeeStoreQuery = `
			INSERT INTO calendar_event_attendee(
			    event_id,
			    user_id,
			    email,
			    status
			)
			VALUES (?, ?, ?, ?)
		`
		sqlReminderReplaceDeleteQuery = `
			DELETE FROM calendar_event_reminder
			WHERE event_id = ?
		`
		sqlReminderStoreQuery = `
			INSERT INTO calendar_event_reminder(
			    event_id,
			    minutes_before,
			    in_app,
			    email
			)
			VALUES (?, ?, ?, ?)
		`
	)

	var e error
	if expected == 0 {
		_, e = r.db.ExecContext(r.ctx, "event.insert", sqlEventInsertQuery, v.ID, v.CalendarID, v.Title, v.Description, v.StartsAt, v.EndsAt, v.Timezone, nullString(v.RRule), v.Version, v.CreatedAt, v.UpdatedAt)
	} else {
		res, err := r.db.ExecContext(r.ctx, "event.update", sqlEventUpdateQuery, v.Title, v.Description, v.StartsAt, v.EndsAt, v.Timezone, nullString(v.RRule), v.UpdatedAt, v.ID, expected)
		e = err
		if e == nil {
			n, _ := res.RowsAffected()
			if n == 0 {
				return domain.Event{}, domain.ErrConflict
			}
			v.Version = expected + 1
		}
	}
	if e != nil {
		return domain.Event{}, e
	}
	if _, e = r.db.ExecContext(r.ctx, "attendee.replace.delete", sqlAttendeeReplaceDeleteQuery, v.ID); e != nil {
		return domain.Event{}, e
	}
	for _, a := range v.Attendees {
		var uid any
		if a.UserID != uuid.Nil() {
			uid = a.UserID
		}
		if _, e = r.db.ExecContext(r.ctx, "attendee.store", sqlAttendeeStoreQuery, v.ID, uid, a.Email, a.Status); e != nil {
			return domain.Event{}, e
		}
	}
	if _, e = r.db.ExecContext(r.ctx, "reminder.replace.delete", sqlReminderReplaceDeleteQuery, v.ID); e != nil {
		return domain.Event{}, e
	}
	for _, m := range v.Reminders {
		if _, e = r.db.ExecContext(r.ctx, "reminder.store", sqlReminderStoreQuery, v.ID, m.MinutesBefore, m.InApp, m.Email); e != nil {
			return domain.Event{}, e
		}
	}
	return v, nil
}

func (r *repo) FindEvent(id uuid.UUID) (domain.Event, error) {
	const sqlEventFindQuery = `
		SELECT
		    id,
		    calendar_id,
		    title,
		    description,
		    starts_at,
		    ends_at,
		    timezone,
		    rrule,
		    version,
		    created_at,
		    updated_at
		FROM calendar_event
		WHERE id = ?
	`

	var v eventRow
	e := r.db.GetContext(r.ctx, "event.find", &v, sqlEventFindQuery, id)
	if errors.Is(e, sql.ErrNoRows) {
		return domain.Event{}, domain.ErrNotFound
	}
	if e != nil {
		return domain.Event{}, e
	}
	return r.expand(v)
}

func (r *repo) ListEvents(cal uuid.UUID, from, to time.Time) ([]domain.Event, error) {
	const (
		allQuery = `
			SELECT
			    id,
			    calendar_id,
			    title,
			    description,
			    starts_at,
			    ends_at,
			    timezone,
			    rrule,
			    version,
			    created_at,
			    updated_at
			FROM calendar_event
			WHERE calendar_id = ?
			ORDER BY starts_at
		`
		fromQuery = `
			SELECT
			    id,
			    calendar_id,
			    title,
			    description,
			    starts_at,
			    ends_at,
			    timezone,
			    rrule,
			    version,
			    created_at,
			    updated_at
			FROM calendar_event
			WHERE calendar_id = ? AND ends_at >= ?
			ORDER BY starts_at
		`
		toQuery = `
			SELECT
			    id,
			    calendar_id,
			    title,
			    description,
			    starts_at,
			    ends_at,
			    timezone,
			    rrule,
			    version,
			    created_at,
			    updated_at
			FROM calendar_event
			WHERE calendar_id = ? AND starts_at <= ?
			ORDER BY starts_at
		`
		fromToQuery = `
			SELECT
			    id,
			    calendar_id,
			    title,
			    description,
			    starts_at,
			    ends_at,
			    timezone,
			    rrule,
			    version,
			    created_at,
			    updated_at
			FROM calendar_event
			WHERE calendar_id = ? AND ends_at >= ? AND starts_at <= ?
			ORDER BY starts_at
		`
	)
	query := allQuery
	args := []any{cal}
	switch {
	case !from.IsZero() && !to.IsZero():
		query = fromToQuery
		args = append(args, from, to)
	case !from.IsZero():
		query = fromQuery
		args = append(args, from)
	case !to.IsZero():
		query = toQuery
		args = append(args, to)
	}
	var rows []eventRow
	if e := r.db.SelectContext(r.ctx, "event.list", &rows, query, args...); e != nil {
		return nil, e
	}
	out := make([]domain.Event, 0, len(rows))
	for _, row := range rows {
		v, e := r.expand(row)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *repo) DeleteEvent(id uuid.UUID) error {
	const sqlEventDeleteQuery = `
		DELETE FROM calendar_event
		WHERE id = ?
	`

	res, e := r.db.ExecContext(r.ctx, "event.delete", sqlEventDeleteQuery, id)
	if e == nil {
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrNotFound
		}
	}
	return e
}

func (r *repo) ReplaceReminderDeliveries(workspace uuid.UUID, event domain.Event) error {
	const (
		sqlReminderDeliveryReplaceQuery = `
			DELETE FROM calendar_reminder_delivery
			WHERE event_id = ? AND delivered_at IS NULL
		`
		sqlReminderDeliveryStoreQuery = `
			INSERT INTO calendar_reminder_delivery(
			    id,
			    event_id,
			    workspace_id,
			    user_id,
			    recipient_email,
			    title,
			    body,
			    rrule,
			    in_app,
			    email,
			    occurrence_starts_at,
			    deliver_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
	)

	if _, e := r.db.ExecContext(r.ctx, "reminder_delivery.replace", sqlReminderDeliveryReplaceQuery, event.ID); e != nil {
		return e
	}
	for _, attendee := range event.Attendees {
		if attendee.Status == "declined" {
			continue
		}
		for _, reminder := range event.Reminders {
			deliverAt := event.StartsAt.Add(-time.Duration(reminder.MinutesBefore) * time.Minute)
			var userID any
			if attendee.UserID != uuid.Nil() {
				userID = attendee.UserID
			}
			_, e := r.db.ExecContext(r.ctx, "reminder_delivery.store", sqlReminderDeliveryStoreQuery, uuid.New(), event.ID, workspace, userID, attendee.Email, event.Title, event.Description, nullString(event.RRule), reminder.InApp, reminder.Email, event.StartsAt, deliverAt)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

func (r *repo) DueReminderDeliveries(now time.Time, limit int) ([]domain.ReminderDelivery, error) {
	const sqlReminderDeliveryDueQuery = `
		SELECT
		    id,
		    event_id,
		    workspace_id,
		    user_id,
		    recipient_email,
		    title,
		    body,
		    rrule,
		    in_app,
		    email,
		    occurrence_starts_at,
		    deliver_at
		FROM calendar_reminder_delivery
		WHERE delivered_at IS NULL AND deliver_at <= ?
		ORDER BY deliver_at
		LIMIT ?
	`

	var rows []struct {
		ID, EventID, WorkspaceID uuid.UUID
		UserID                   *uuid.UUID     `db:"user_id"`
		RecipientEmail           string         `db:"recipient_email"`
		Title                    string         `db:"title"`
		Body                     string         `db:"body"`
		RRule                    sql.NullString `db:"rrule"`
		InApp                    bool           `db:"in_app"`
		Email                    bool           `db:"email"`
		OccurrenceStartsAt       time.Time      `db:"occurrence_starts_at"`
		DeliverAt                time.Time      `db:"deliver_at"`
	}
	e := r.db.SelectContext(r.ctx, "reminder_delivery.due", &rows, sqlReminderDeliveryDueQuery, now, limit)
	if e != nil {
		return nil, e
	}
	out := make([]domain.ReminderDelivery, 0, len(rows))
	for _, row := range rows {
		v := domain.ReminderDelivery{ID: row.ID, EventID: row.EventID, WorkspaceID: row.WorkspaceID, RecipientEmail: row.RecipientEmail, Title: row.Title, Body: row.Body, RRule: row.RRule.String, InApp: row.InApp, Email: row.Email, OccurrenceStartsAt: row.OccurrenceStartsAt, DeliverAt: row.DeliverAt}
		if row.UserID != nil {
			v.UserID = *row.UserID
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *repo) CompleteReminderDelivery(id uuid.UUID, next time.Time) error {
	const (
		sqlReminderDeliveryCompleteQuery = `
			UPDATE calendar_reminder_delivery
			SET
			    delivered_at = UTC_TIMESTAMP(6)
			WHERE id = ?
		`
		sqlReminderDeliveryRescheduleQuery = `
			UPDATE calendar_reminder_delivery
			SET
			    occurrence_starts_at = DATE_ADD(occurrence_starts_at, INTERVAL TIMESTAMPDIFF(SECOND, deliver_at, ?) SECOND),
			    deliver_at = ?
			WHERE id = ?
		`
	)

	if next.IsZero() {
		_, e := r.db.ExecContext(r.ctx, "reminder_delivery.complete", sqlReminderDeliveryCompleteQuery, id)
		return e
	}
	_, e := r.db.ExecContext(r.ctx, "reminder_delivery.reschedule", sqlReminderDeliveryRescheduleQuery, next, next, id)
	return e
}

func (r *repo) expand(v eventRow) (domain.Event, error) {
	const (
		sqlAttendeeListQuery = `
			SELECT
			    user_id,
			    email,
			    status
			FROM calendar_event_attendee
			WHERE event_id = ?
			ORDER BY email
		`
		sqlReminderListQuery = `
			SELECT
			    minutes_before,
			    in_app,
			    email
			FROM calendar_event_reminder
			WHERE event_id = ?
			ORDER BY minutes_before DESC
		`
	)

	out := domain.Event{ID: v.ID, CalendarID: v.CalendarID, Title: v.Title, Description: v.Description, StartsAt: v.StartsAt, EndsAt: v.EndsAt, Timezone: v.Timezone, RRule: v.RRule.String, Version: v.Version, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	var aa []struct {
		UserID *uuid.UUID `db:"user_id"`
		Email  string     `db:"email"`
		Status string     `db:"status"`
	}
	if e := r.db.SelectContext(r.ctx, "attendee.list", &aa, sqlAttendeeListQuery, v.ID); e != nil {
		return domain.Event{}, e
	}
	for _, a := range aa {
		x := domain.Attendee{Email: a.Email, Status: a.Status}
		if a.UserID != nil {
			x.UserID = *a.UserID
		}
		out.Attendees = append(out.Attendees, x)
	}
	var rr []struct {
		Minutes int  `db:"minutes_before"`
		InApp   bool `db:"in_app"`
		Email   bool `db:"email"`
	}
	if e := r.db.SelectContext(r.ctx, "reminder.list", &rr, sqlReminderListQuery, v.ID); e != nil {
		return domain.Event{}, e
	}
	for _, m := range rr {
		out.Reminders = append(out.Reminders, domain.Reminder{MinutesBefore: m.Minutes, InApp: m.InApp, Email: m.Email})
	}
	return out, nil
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
