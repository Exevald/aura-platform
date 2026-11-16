package domain

import (
	events "planner/api/events"
	"strings"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"
)

type Service struct {
	repo   Repository
	events event.Dispatcher
}

func NewService(repository Repository, dispatcher event.Dispatcher) *Service {
	return &Service{repository, dispatcher}
}

func (s *Service) CreateProject(w, a uuid.UUID, name string) (Project, error) {
	if w == uuid.Nil() || a == uuid.Nil() || !validName(name, 120) {
		return Project{}, ErrInvalid
	}
	now := time.Now().UTC()
	v := Project{ID: uuid.New(), WorkspaceID: w, Name: strings.TrimSpace(name), CreatedAt: now, UpdatedAt: now}
	if e := s.repo.StoreProject(v); e != nil {
		return Project{}, e
	}
	e := s.events.Dispatch(&events.ProjectChanged{
		ProjectId:      v.ID.String(),
		WorkspaceId:    w.String(),
		ActorUserId:    a.String(),
		Action:         "created",
		OccurredAtUnix: now.Unix(),
	})
	return v, e
}

func (s *Service) ListProjects(w uuid.UUID) ([]Project, error) {
	if w == uuid.Nil() {
		return nil, ErrInvalid
	}
	return s.repo.ListProjects(w)
}

func (s *Service) GetProject(w, id uuid.UUID) (Project, error) {
	v, e := s.repo.FindProject(id)
	if e != nil || v.WorkspaceID != w {
		return Project{}, ErrNotFound
	}
	return v, nil
}

func (s *Service) UpdateProject(w, a, id uuid.UUID, name string) (Project, error) {
	v, e := s.GetProject(w, id)
	if e != nil {
		return Project{}, e
	}
	if !validName(name, 120) {
		return Project{}, ErrInvalid
	}
	v.Name, v.UpdatedAt = strings.TrimSpace(name), time.Now().UTC()
	if e = s.repo.StoreProject(v); e != nil {
		return Project{}, e
	}
	e = s.events.Dispatch(&events.ProjectChanged{
		ProjectId:      id.String(),
		WorkspaceId:    w.String(),
		ActorUserId:    a.String(),
		Action:         "updated",
		OccurredAtUnix: v.UpdatedAt.Unix(),
	})
	return v, e
}

func (s *Service) DeleteProject(w, a, id uuid.UUID) error {
	if _, e := s.GetProject(w, id); e != nil {
		return e
	}
	if e := s.repo.DeleteProject(w, id); e != nil {
		return e
	}
	now := time.Now().UTC()
	return s.events.Dispatch(&events.ProjectChanged{
		ProjectId:      id.String(),
		WorkspaceId:    w.String(),
		ActorUserId:    a.String(),
		Action:         "deleted",
		OccurredAtUnix: now.Unix(),
	})
}

func (s *Service) CreateTask(w, a, p uuid.UUID, v Task) (Task, error) {
	if _, e := s.GetProject(w, p); e != nil {
		return Task{}, e
	}
	if !validTask(v) {
		return Task{}, ErrInvalid
	}
	now := time.Now().UTC()
	if v.ID == uuid.Nil() {
		v.ID = uuid.New()
	}
	v.ProjectID, v.Version, v.CreatedAt, v.UpdatedAt = p, 1, now, now
	stored, e := s.repo.StoreTask(v, 0)
	if e != nil {
		return Task{}, e
	}
	e = s.events.Dispatch(taskEvent(stored, w, a, "created"))
	return stored, e
}

func (s *Service) ListTasks(w, p uuid.UUID) ([]Task, error) {
	if _, e := s.GetProject(w, p); e != nil {
		return nil, e
	}
	return s.repo.ListTasks(p)
}

func (s *Service) GetTask(w, id uuid.UUID) (Task, error) {
	v, e := s.repo.FindTask(id)
	if e != nil {
		return Task{}, ErrNotFound
	}
	if _, e = s.GetProject(w, v.ProjectID); e != nil {
		return Task{}, e
	}
	return v, nil
}

func (s *Service) UpdateTask(w, a, id uuid.UUID, v Task) (Task, error) {
	old, e := s.GetTask(w, id)
	if e != nil {
		return Task{}, e
	}
	if v.Version != old.Version {
		return Task{}, ErrConflict
	}
	if !validTask(v) {
		return Task{}, ErrInvalid
	}
	v.ID, v.ProjectID, v.CreatedAt, v.UpdatedAt = id, old.ProjectID, old.CreatedAt, time.Now().UTC()
	stored, e := s.repo.StoreTask(v, old.Version)
	if e != nil {
		return Task{}, e
	}
	e = s.events.Dispatch(taskEvent(stored, w, a, "updated"))
	return stored, e
}

func (s *Service) DeleteTask(w, a, id uuid.UUID) (Task, error) {
	v, e := s.GetTask(w, id)
	if e != nil {
		return Task{}, e
	}
	if e = s.repo.DeleteTask(id); e != nil {
		return Task{}, e
	}
	if e = s.events.Dispatch(taskEvent(v, w, a, "deleted")); e != nil {
		return Task{}, e
	}
	return v, nil
}

func taskEvent(v Task, w, a uuid.UUID, action string) *events.TaskChanged {
	return &events.TaskChanged{
		TaskId:         v.ID.String(),
		ProjectId:      v.ProjectID.String(),
		WorkspaceId:    w.String(),
		ActorUserId:    a.String(),
		Action:         action,
		AssigneeUserId: v.AssigneeUserID.String(),
		OccurredAtUnix: time.Now().UTC().Unix(),
	}
}

func validName(v string, maxLength int) bool {
	return strings.TrimSpace(v) != "" && len([]rune(v)) <= maxLength
}

func validTask(v Task) bool {
	if !validName(v.Title, 255) {
		return false
	}
	if v.Status != "todo" && v.Status != "in_progress" && v.Status != "done" {
		return false
	}
	if v.Priority != "low" && v.Priority != "normal" && v.Priority != "high" {
		return false
	}
	if v.SyncToCalendar && (v.DueAt == nil || v.CalendarID == uuid.Nil() || v.Timezone == "") {
		return false
	}
	return true
}
