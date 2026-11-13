package domain

import (
	"testing"
	"time"
	"uuid"
)

type mailRepositoryStub struct {
	id  uuid.UUID
	job Job
}

func (r *mailRepositoryStub) NextID() (uuid.UUID, error) {
	return r.id, nil
}

func (r *mailRepositoryStub) Create(job Job) (Job, bool, error) {
	r.job = job
	return job, true, nil
}

func (r *mailRepositoryStub) Find(uuid.UUID) (Job, error) {
	return r.job, nil
}

func (r *mailRepositoryStub) FindByKey(string) (Job, error) {
	return r.job, nil
}

func (r *mailRepositoryStub) StoreResult(job Job) error {
	r.job = job
	return nil
}

func TestCreateUsesRepositoryIDAndSuppliedTime(t *testing.T) {
	id := uuid.New()
	now := time.Date(2026, time.November, 13, 12, 0, 0, 0, time.UTC)
	repo := &mailRepositoryStub{id: id}

	job, created, err := NewService(repo, nil).Create("event-1", "recipient@example.com", "subject", "body", now)
	if err != nil || !created {
		t.Fatalf("create mail: created=%v, err=%v", created, err)
	}
	if job.ID != id || !job.CreatedAt.Equal(now) || !job.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected job identity or time: %+v", job)
	}
}
