package domain

import (
	events "filestorage/api/events"
	"path/filepath"
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

func (s *Service) Store(v File, now time.Time) (File, error) {
	v.FileName = filepath.Base(strings.TrimSpace(v.FileName))
	if v.ID == uuid.Nil() || v.WorkspaceID == uuid.Nil() || v.OwnerUserID == uuid.Nil() || v.FileName == "" || v.SizeBytes < 0 || len(v.ChecksumSHA256) != 64 {
		return File{}, ErrInvalid
	}
	v.CreatedAt = now.UTC()
	if e := s.repo.Store(v); e != nil {
		return File{}, e
	}
	e := s.events.Dispatch(&events.FileStored{FileId: v.ID.String(), WorkspaceId: v.WorkspaceID.String(), OwnerUserId: v.OwnerUserID.String(), FileName: v.FileName, SizeBytes: v.SizeBytes, OccurredAtUnix: v.CreatedAt.Unix()})
	return v, e
}

func (s *Service) Get(w, id uuid.UUID) (File, error) {
	v, e := s.repo.Find(id)
	if e != nil || v.WorkspaceID != w {
		return File{}, ErrNotFound
	}
	return v, nil
}

func (s *Service) List(w uuid.UUID) ([]File, error) {
	if w == uuid.Nil() {
		return nil, ErrInvalid
	}
	return s.repo.List(w)
}

func (s *Service) Delete(w, actor, id uuid.UUID, now time.Time) (File, error) {
	v, e := s.Get(w, id)
	if e != nil {
		return File{}, e
	}
	if v.OwnerUserID != actor {
		return File{}, ErrForbidden
	}
	if e = s.repo.Delete(id); e != nil {
		return File{}, e
	}
	e = s.events.Dispatch(&events.FileDeleted{FileId: id.String(), WorkspaceId: w.String(), ActorUserId: actor.String(), OccurredAtUnix: now.Unix()})
	return v, e
}
