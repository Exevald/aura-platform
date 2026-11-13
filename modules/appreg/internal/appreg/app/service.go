package app

import "uuid"

type Service struct {
	apps []App
}

func NewService(apps []App) *Service {
	return &Service{apps: append([]App(nil), apps...)}
}

func (s *Service) List(defaultOnly bool) []App {
	return Apps(s.apps, defaultOnly)
}

func (s *Service) Get(id uuid.UUID) (App, error) {
	return Find(s.apps, id)
}
