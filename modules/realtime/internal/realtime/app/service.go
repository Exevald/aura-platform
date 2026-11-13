package app

import (
	"uuid"
)

type Service struct {
	hub *Hub
}

func NewService() *Service {
	return &Service{hub: NewHub()}
}

func (s *Service) Subscribe(user, workspace uuid.UUID) (subscription *Subscription, unsubscribe func()) {
	return s.hub.Subscribe(user, workspace)
}

func (s *Service) Publish(user, workspace uuid.UUID, t string, p []byte) int {
	return s.hub.Publish(user, workspace, Message{Type: t, Payload: p})
}
