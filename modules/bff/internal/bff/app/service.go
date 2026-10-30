package app

import (
	"context"
	"uuid"
)

type Profile struct {
	UserID      uuid.UUID
	DisplayName string
	FirstName   string
	LastName    string
	Email       string
}

type UserClient interface {
	Get(context.Context, uuid.UUID) (Profile, error)
	Update(context.Context, Profile) error
}

type Service struct{ users UserClient }

func NewService(users UserClient) *Service {
	return &Service{users: users}
}

func (s *Service) GetMe(ctx context.Context, id uuid.UUID) (Profile, error) {
	return s.users.Get(ctx, id)
}

func (s *Service) UpdateMe(ctx context.Context, profile Profile) (Profile, error) {
	if err := s.users.Update(ctx, profile); err != nil {
		return Profile{}, err
	}
	return s.users.Get(ctx, profile.UserID)
}

func (s *Service) Echo(context.Context, string) (uuid.UUID, error) { return uuid.New(), nil }
