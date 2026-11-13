package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"filestorage/internal/filestorage/domain"
	"fmt"
	"io"
	"time"
	"uuid"

	domainevent "github.com/distributed-programming-2026/lib/event"
)

type RepositoryProvider interface {
	File(context.Context) domain.Repository
}

type UnitOfWork interface {
	Execute(context.Context, func(RepositoryProvider) error) error
}

type DispatcherFactory interface {
	NewDispatcher(context.Context) domainevent.Dispatcher
}

type ObjectStore interface {
	Put(context.Context, string, io.Reader, int64, string) (int64, error)
	Get(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

type Service struct {
	unit    UnitOfWork
	events  DispatcherFactory
	objects ObjectStore
}

func NewService(u UnitOfWork, e DispatcherFactory, stores ...ObjectStore) *Service {
	s := &Service{unit: u, events: e}
	if len(stores) > 0 {
		s.objects = stores[0]
	}
	return s
}

func (s *Service) domain(ctx context.Context, p RepositoryProvider) *domain.Service {
	return domain.NewService(p.File(ctx), s.events.NewDispatcher(ctx))
}

func (s *Service) Upload(ctx context.Context, w, u uuid.UUID, name, contentType string, size int64, body io.Reader) (out domain.File, err error) {
	if s.objects == nil {
		return domain.File{}, fmt.Errorf("object store unavailable")
	}
	id := uuid.New()
	key := w.String() + "/" + id.String()
	hash := sha256.New()
	actual, e := s.objects.Put(ctx, key, io.TeeReader(body, hash), size, contentType)
	if e != nil {
		return domain.File{}, e
	}
	value := domain.File{ID: id, WorkspaceID: w, OwnerUserID: u, ObjectKey: key, FileName: name, ContentType: contentType, SizeBytes: actual, ChecksumSHA256: hex.EncodeToString(hash.Sum(nil))}
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error {
		out, err = s.domain(ctx, p).Store(value, time.Now().UTC())
		return err
	})
	if err != nil {
		_ = s.objects.Delete(context.WithoutCancel(ctx), key)
	}
	return
}

func (s *Service) Get(ctx context.Context, w, id uuid.UUID) (out domain.File, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error { out, err = s.domain(ctx, p).Get(w, id); return err })
	return
}

func (s *Service) Download(ctx context.Context, w, id uuid.UUID) (domain.File, io.ReadCloser, error) {
	v, e := s.Get(ctx, w, id)
	if e != nil {
		return domain.File{}, nil, e
	}
	body, e := s.objects.Get(ctx, v.ObjectKey)
	return v, body, e
}

func (s *Service) List(ctx context.Context, w uuid.UUID) (out []domain.File, err error) {
	err = s.unit.Execute(ctx, func(p RepositoryProvider) error { out, err = s.domain(ctx, p).List(w); return err })
	return
}

func (s *Service) Delete(ctx context.Context, w, u, id uuid.UUID) error {
	var value domain.File
	err := s.unit.Execute(ctx, func(p RepositoryProvider) error {
		var e error
		value, e = s.domain(ctx, p).Get(w, id)
		if e != nil {
			return e
		}
		if value.OwnerUserID != u {
			return domain.ErrForbidden
		}
		return nil
	})
	if err != nil {
		return err
	}
	if err = s.objects.Delete(ctx, value.ObjectKey); err != nil {
		return err
	}
	return s.unit.Execute(ctx, func(p RepositoryProvider) error {
		_, e := s.domain(ctx, p).Delete(w, u, id, time.Now().UTC())
		return e
	})
}
