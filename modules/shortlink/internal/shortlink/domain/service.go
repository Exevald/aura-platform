package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"net/url"
	"time"
	"uuid"
)

const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type LinkService struct{ repo LinkRepository }

func NewLinkService(repo LinkRepository) *LinkService { return &LinkService{repo: repo} }

func (s *LinkService) Create(ownerID uuid.UUID, target string, now time.Time) (Link, error) {
	if ownerID == uuid.Nil() || !validTarget(target) {
		return Link{}, ErrInvalidTarget
	}
	id, err := s.repo.NextID()
	if err != nil {
		return Link{}, err
	}
	for range 5 {
		code, codeErr := randomCode(10)
		if codeErr != nil {
			return Link{}, codeErr
		}
		link := Link{ID: id, OwnerID: ownerID, Code: code, TargetURL: target, CreatedAt: now.UTC()}
		if err = s.repo.Store(link); errors.Is(err, ErrCodeConflict) {
			continue
		} else if err != nil {
			return Link{}, err
		}
		return link, nil
	}
	return Link{}, ErrCodeConflict
}

func (s *LinkService) Ensure(ownerID uuid.UUID, key, target string, now time.Time) (Link, error) {
	if key == "" || ownerID == uuid.Nil() || !validTarget(target) {
		return Link{}, ErrInvalidTarget
	}
	id := LinkIDForKey(ownerID, key)
	if existing, err := s.repo.FindByID(id); err == nil {
		if existing.OwnerID != ownerID || existing.TargetURL != target {
			return Link{}, ErrInvalidTarget
		}
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Link{}, err
	}
	for range 5 {
		code, err := randomCode(10)
		if err != nil {
			return Link{}, err
		}
		link := Link{ID: id, OwnerID: ownerID, Code: code, TargetURL: target, CreatedAt: now.UTC()}
		err = s.repo.Store(link)
		if err == nil {
			return link, nil
		}
		if !errors.Is(err, ErrCodeConflict) {
			return Link{}, err
		}
		if existing, e := s.repo.FindByID(id); e == nil {
			return existing, nil
		}
	}
	return Link{}, ErrCodeConflict
}

func LinkIDForKey(ownerID uuid.UUID, key string) uuid.UUID {
	hash := sha256.Sum256([]byte(ownerID.String() + ":" + key))
	var id uuid.UUID
	copy(id[:], hash[:16])
	id[6] = id[6]&0x0f | 0x50
	id[8] = id[8]&0x3f | 0x80
	return id
}

func (s *LinkService) List(ownerID uuid.UUID) ([]Link, error) {
	return s.repo.ListByOwner(ownerID)
}

func (s *LinkService) Resolve(code string) (Link, error) {
	return s.repo.FindByCode(code)
}

func (s *LinkService) Delete(ownerID, id uuid.UUID) error {
	return s.repo.Delete(ownerID, id)
}

func validTarget(target string) bool {
	if target == "" || len(target) > 2048 {
		return false
	}
	parsed, err := url.ParseRequestURI(target)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func randomCode(length int) (string, error) {
	buffer := make([]byte, length)
	for i := range buffer {
		var value [1]byte
		for {
			if _, err := rand.Read(value[:]); err != nil {
				return "", err
			}
			if int(value[0]) < 256-256%len(alphabet) {
				buffer[i] = alphabet[int(value[0])%len(alphabet)]
				break
			}
		}
	}
	return string(buffer), nil
}
