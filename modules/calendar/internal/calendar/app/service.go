package app

import (
	"calendar/internal/calendar/domain"
	"context"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	Calendar(context.Context) domain.Repository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type NotificationSender interface {
	SendReminder(context.Context, domain.ReminderDelivery) error
}

type Service struct {
	unit          UnitOfWork
	events        DispatcherFactory
	notifications NotificationSender
}

func NewService(u UnitOfWork, e DispatcherFactory, notifications NotificationSender) *Service {
	return &Service{unit: u, events: e, notifications: notifications}
}

func (s *Service) run(ctx context.Context, fn func(*domain.Service) error) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return fn(domain.NewService(p.Calendar(ctx), s.events.NewDispatcher(ctx)))
	})
}

func (s *Service) CreateCalendar(ctx context.Context, w, a uuid.UUID, n, tz string) (out domain.Calendar, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateCalendar(w, a, n, tz)
		return err
	})
	return
}

func (s *Service) ListCalendars(ctx context.Context, w uuid.UUID) (out []domain.Calendar, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.List(w)
		return err
	})
	return
}

func (s *Service) GetCalendar(ctx context.Context, w, id uuid.UUID) (out domain.Calendar, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.Get(w, id)
		return err
	})
	return
}

func (s *Service) UpdateCalendar(ctx context.Context, w, a, id uuid.UUID, n, tz string) (out domain.Calendar, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.UpdateCalendar(w, a, id, n, tz)
		return err
	})
	return
}

func (s *Service) DeleteCalendar(ctx context.Context, w, a, id uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.DeleteCalendar(w, a, id)
	})
}

func (s *Service) CreateEvent(ctx context.Context, w, a, c uuid.UUID, v domain.Event) (out domain.Event, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateEvent(w, a, c, v)
		return err
	})
	return
}

func (s *Service) UpsertTaskEvent(ctx context.Context, w, a, id, c uuid.UUID, v domain.Event) (out domain.Event, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateEventWithID(w, a, id, c, v)
		return err
	})
	return
}

func (s *Service) ListEvents(ctx context.Context, w, c uuid.UUID, from, to time.Time) (out []domain.Event, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.ListEvents(w, c, from, to)
		return err
	})
	return
}

func (s *Service) UpdateEvent(ctx context.Context, w, a, id uuid.UUID, v domain.Event) (out domain.Event, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.UpdateEvent(w, a, id, v)
		return err
	})
	return
}

func (s *Service) DeleteEvent(ctx context.Context, w, a, id uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.DeleteEvent(w, a, id)
	})
}

func (s *Service) DeliverDueReminders(ctx context.Context, now time.Time, limit int) error {
	var due []domain.ReminderDelivery
	if err := s.run(ctx, func(d *domain.Service) error { var err error; due, err = d.DueReminders(now, limit); return err }); err != nil {
		return err
	}
	for _, reminder := range due {
		if s.notifications == nil {
			continue
		}
		if err := s.notifications.SendReminder(ctx, reminder); err != nil {
			continue
		}
		if err := s.run(ctx, func(d *domain.Service) error {
			return d.CompleteReminder(reminder)
		}); err != nil {
			return err
		}
	}
	return nil
}
