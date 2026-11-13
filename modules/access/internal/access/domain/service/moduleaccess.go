package service

import (
	"access/internal/access/domain/model"
	"time"
	"uuid"

	"github.com/distributed-programming-2026/lib/event"

	accessevents "access/api/events"
)

type ModuleAccessService interface {
	Create(userID, moduleID uuid.UUID, effect model.Effect) (uuid.UUID, error)
	CreateWithSource(userID, moduleID uuid.UUID, effect model.Effect, sourceID string) (uuid.UUID, error)
	CreateWithID(id, userID, moduleID uuid.UUID, effect model.Effect, sourceID string) error
	Update(id, userID, moduleID uuid.UUID, effect model.Effect) error
	Delete(id uuid.UUID) error
	CheckModuleAccess(userID, moduleID uuid.UUID) (bool, error)
}

type moduleAccessService struct {
	repo       model.ModuleAccessRepository
	dispatcher event.Dispatcher
}

func NewModuleAccessService(
	repo model.ModuleAccessRepository,
	dispatcher event.Dispatcher,
) ModuleAccessService {
	return &moduleAccessService{
		repo:       repo,
		dispatcher: dispatcher,
	}
}

func (s *moduleAccessService) Create(userID, moduleID uuid.UUID, effect model.Effect) (uuid.UUID, error) {
	return s.CreateWithSource(userID, moduleID, effect, "direct")
}

func (s *moduleAccessService) CreateWithSource(userID, moduleID uuid.UUID, effect model.Effect, sourceID string) (uuid.UUID, error) {
	if err := validateModuleAccess(userID, moduleID, effect); err != nil {
		return uuid.Nil(), err
	}
	id, err := s.repo.NextID()
	if err != nil {
		return uuid.Nil(), err
	}
	if err = s.CreateWithID(id, userID, moduleID, effect, sourceID); err != nil {
		return uuid.Nil(), err
	}
	return id, nil
}

func (s *moduleAccessService) CreateWithID(id, userID, moduleID uuid.UUID, effect model.Effect, sourceID string) error {
	if id == uuid.Nil() || sourceID == "" {
		return model.ErrInvalidModuleAccess
	}
	if err := validateModuleAccess(userID, moduleID, effect); err != nil {
		return err
	}
	now := time.Now().UTC()
	access := model.ModuleAccess{
		ID:        id,
		UserID:    userID,
		ModuleID:  moduleID,
		Effect:    effect,
		SourceID:  sourceID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Store(access); err != nil {
		return err
	}
	if err := s.dispatch("created", access, now); err != nil {
		return err
	}
	return nil
}

func (s *moduleAccessService) Update(id, userID, moduleID uuid.UUID, effect model.Effect) error {
	if err := validateModuleAccess(userID, moduleID, effect); err != nil {
		return err
	}
	access, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	access.UserID, access.ModuleID, access.Effect = userID, moduleID, effect
	access.UpdatedAt = time.Now().UTC()
	if err = s.repo.Store(access); err != nil {
		return err
	}
	return s.dispatch("updated", access, access.UpdatedAt)
}

func (s *moduleAccessService) Delete(id uuid.UUID) error {
	access, err := s.repo.Find(id)
	if err != nil {
		return err
	}
	if err = s.repo.Delete(id); err != nil {
		return err
	}
	return s.dispatch("deleted", access, time.Now().UTC())
}

func (s *moduleAccessService) CheckModuleAccess(userID, moduleID uuid.UUID) (bool, error) {
	accesses, err := s.repo.FindByUserAndModule(userID, moduleID)
	if err != nil {
		return false, err
	}
	allowed := false
	for _, access := range accesses {
		if access.Effect == model.EffectDeny {
			return false, nil
		}
		allowed = allowed || access.Effect == model.EffectAllow
	}
	return allowed, nil
}

func (s *moduleAccessService) dispatch(kind string, access model.ModuleAccess, at time.Time) error {
	return s.dispatcher.Dispatch(&accessevents.ModuleAccessChanged{
		Type:           kind,
		Id:             access.ID.String(),
		UserId:         access.UserID.String(),
		ModuleId:       access.ModuleID.String(),
		Effect:         string(access.Effect),
		OccurredAtUnix: at.Unix(),
	})
}

func validateModuleAccess(userID, moduleID uuid.UUID, effect model.Effect) error {
	if userID == uuid.Nil() || moduleID == uuid.Nil() || (effect != model.EffectAllow && effect != model.EffectDeny) {
		return model.ErrInvalidModuleAccess
	}
	return nil
}
