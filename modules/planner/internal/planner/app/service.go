package app

import (
	"context"
	"planner/internal/planner/domain"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	Planner(context.Context) domain.Repository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type CalendarAdapter interface {
	UpsertTaskEvent(context.Context, uuid.UUID, uuid.UUID, domain.Task) (uuid.UUID, error)
	DeleteTaskEvent(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type Service struct {
	unit     UnitOfWork
	events   DispatcherFactory
	calendar CalendarAdapter
}

func NewService(u UnitOfWork, e DispatcherFactory, calendars ...CalendarAdapter) *Service {
	s := &Service{unit: u, events: e}
	if len(calendars) > 0 {
		s.calendar = calendars[0]
	}
	return s
}

func (s *Service) run(ctx context.Context, fn func(*domain.Service) error) error {
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		return fn(domain.NewService(p.Planner(ctx), s.events.NewDispatcher(ctx)))
	})
}

func (s *Service) CreateProject(ctx context.Context, w, a uuid.UUID, n string) (out domain.Project, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateProject(w, a, n)
		return err
	})
	return
}

func (s *Service) ListProjects(ctx context.Context, w uuid.UUID) (out []domain.Project, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.ListProjects(w)
		return err
	})
	return
}

func (s *Service) UpdateProject(ctx context.Context, w, a, id uuid.UUID, n string) (out domain.Project, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.UpdateProject(w, a, id, n)
		return err
	})
	return
}

func (s *Service) DeleteProject(ctx context.Context, w, a, id uuid.UUID) error {
	return s.run(ctx, func(d *domain.Service) error {
		return d.DeleteProject(w, a, id)
	})
}

func (s *Service) CreateTask(ctx context.Context, w, a, p uuid.UUID, v domain.Task) (out domain.Task, err error) {
	v.ID = uuid.New()
	if v.SyncToCalendar && s.calendar != nil {
		v.CalendarEventID, err = s.calendar.UpsertTaskEvent(ctx, w, a, v)
		if err != nil {
			return domain.Task{}, err
		}
	}
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.CreateTask(w, a, p, v)
		return err
	})
	if err != nil && v.CalendarEventID != uuid.Nil() && s.calendar != nil {
		_ = s.calendar.DeleteTaskEvent(context.WithoutCancel(ctx), w, a, v.CalendarEventID)
	}
	return
}

func (s *Service) ListTasks(ctx context.Context, w, p uuid.UUID) (out []domain.Task, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.ListTasks(w, p)
		return err
	})
	return
}

func (s *Service) GetTask(ctx context.Context, w, id uuid.UUID) (out domain.Task, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.GetTask(w, id)
		return err
	})
	return
}

func (s *Service) UpdateTask(ctx context.Context, w, a, id uuid.UUID, v domain.Task) (out domain.Task, err error) {
	old, err := s.GetTask(ctx, w, id)
	if err != nil {
		return domain.Task{}, err
	}
	if old.Version != v.Version {
		return domain.Task{}, domain.ErrConflict
	}
	v.ID = id
	if v.SyncToCalendar && s.calendar != nil {
		v.CalendarEventID = old.CalendarEventID
		if v.CalendarEventID == uuid.Nil() {
			v.CalendarEventID = id
		}
		v.CalendarEventID, err = s.calendar.UpsertTaskEvent(ctx, w, a, v)
		if err != nil {
			return domain.Task{}, err
		}
	}
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.UpdateTask(w, a, id, v)
		return err
	})
	if err != nil {
		return domain.Task{}, err
	}
	if !v.SyncToCalendar && old.CalendarEventID != uuid.Nil() && s.calendar != nil {
		err = s.calendar.DeleteTaskEvent(ctx, w, a, old.CalendarEventID)
	}
	return out, err
}

func (s *Service) DeleteTask(ctx context.Context, w, a, id uuid.UUID) (out domain.Task, err error) {
	err = s.run(ctx, func(d *domain.Service) error {
		out, err = d.DeleteTask(w, a, id)
		return err
	})
	if err == nil && out.CalendarEventID != uuid.Nil() && s.calendar != nil {
		err = s.calendar.DeleteTaskEvent(ctx, w, a, out.CalendarEventID)
	}
	return
}

func TaskEventTimes(v domain.Task) (start, end time.Time) {
	if v.DueAt == nil {
		return time.Time{}, time.Time{}
	}
	return *v.DueAt, v.DueAt.Add(time.Hour)
}
