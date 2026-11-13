package domain

import (
	"time"
	"uuid"
)

type GroupMembership struct {
	GroupID   uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
}

type GroupMembershipRepository interface {
	ListByGroup(uuid.UUID) ([]GroupMembership, error)
	Store(GroupMembership) (bool, error)
	Delete(groupID, userID uuid.UUID) (bool, error)
}
