package domain

import (
	"errors"
	"time"
	"uuid"
)

var (
	ErrInvalid  = errors.New("invalid planner data")
	ErrNotFound = errors.New("planner resource not found")
	ErrConflict = errors.New("planner resource version conflict")
)

type Project struct {
	ID, WorkspaceID      uuid.UUID
	Name                 string
	CreatedAt, UpdatedAt time.Time
}

type Task struct {
	ID, ProjectID, AssigneeUserID, CalendarID, CalendarEventID uuid.UUID
	Title, Description, Status, Priority, Timezone             string
	DueAt                                                      *time.Time
	SyncToCalendar                                             bool
	Version                                                    int64
	CreatedAt, UpdatedAt                                       time.Time
}

type Repository interface {
	StoreProject(Project) error
	FindProject(uuid.UUID) (Project, error)
	ListProjects(uuid.UUID) ([]Project, error)
	DeleteProject(uuid.UUID, uuid.UUID) error
	StoreTask(Task, int64) (Task, error)
	FindTask(uuid.UUID) (Task, error)
	ListTasks(uuid.UUID) ([]Task, error)
	DeleteTask(uuid.UUID) error
}
