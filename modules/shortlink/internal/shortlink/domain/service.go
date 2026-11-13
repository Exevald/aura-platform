package domain

import (
	"crypto/rand"
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
