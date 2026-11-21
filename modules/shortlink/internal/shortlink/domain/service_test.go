package domain

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
)

type memoryLinks struct{ values []Link }

func (*memoryLinks) NextID() (uuid.UUID, error) {
	return uuid.New(), nil
}

func (r *memoryLinks) Store(link Link) error {
	for _, value := range r.values {
		if value.Code == link.Code {
			return ErrCodeConflict
		}
	}
	r.values = append(r.values, link)
	return nil
}

func (r *memoryLinks) ListByOwner(owner uuid.UUID) (out []Link, _ error) {
	for _, value := range r.values {
		if value.OwnerID == owner {
			out = append(out, value)
		}
	}
	return
}

func (r *memoryLinks) FindByCode(code string) (Link, error) {
	for _, value := range r.values {
		if value.Code == code {
			return value, nil
		}
	}
	return Link{}, ErrNotFound
}

func (r *memoryLinks) FindByID(id uuid.UUID) (Link, error) {
	for _, value := range r.values {
		if value.ID == id {
			return value, nil
		}
	}
	return Link{}, ErrNotFound
}

func (r *memoryLinks) Delete(owner, id uuid.UUID) error {
	for index, value := range r.values {
		if value.OwnerID == owner && value.ID == id {
			r.values = append(r.values[:index], r.values[index+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func TestCreateResolveAndDeleteLink(t *testing.T) {
	repo := &memoryLinks{}
	svc := NewLinkService(repo)
	owner := uuid.New()
	now := time.Date(2026, time.November, 13, 12, 0, 0, 0, time.UTC)
	link, err := svc.Create(owner, "https://example.com/path?q=1", now)
	require.NoError(t, err)
	require.Len(t, link.Code, 10)
	require.Equal(t, now, link.CreatedAt)
	resolved, err := svc.Resolve(link.Code)
	require.NoError(t, err)
	require.Equal(t, link, resolved)
	require.ErrorIs(t, svc.Delete(uuid.New(), link.ID), ErrNotFound)
	require.NoError(t, svc.Delete(owner, link.ID))
	_, err = svc.Resolve(link.Code)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestRejectsUnsafeTargets(t *testing.T) {
	svc := NewLinkService(&memoryLinks{})
	for _, target := range []string{"", "/relative", "javascript:alert(1)", "ftp://example.com"} {
		_, err := svc.Create(uuid.New(), target, time.Time{})
		require.ErrorIs(t, err, ErrInvalidTarget)
	}
}

func TestEnsureIsIdempotent(t *testing.T) {
	repo := &memoryLinks{}
	svc := NewLinkService(repo)
	owner := uuid.New()
	first, err := svc.Ensure(owner, "publication-1", "http://aura-platform.home/p/example", time.Now())
	require.NoError(t, err)
	second, err := svc.Ensure(owner, "publication-1", first.TargetURL, time.Now())
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Len(t, repo.values, 1)
	require.NoError(t, svc.Delete(owner, first.ID))
}
