package app

import "uuid"

type Message struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type Subscription struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	WorkspaceID uuid.UUID
	Messages    chan Message
}
