package public

import (
	api "realtime/api/server/publicapi"
	"realtime/internal/realtime/app"
)

type handler struct{}

func NewHandler(*app.Service) api.StrictServerInterface { return &handler{} }
