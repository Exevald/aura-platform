package public

import (
	api "mailer/api/server/publicapi"
	"mailer/internal/mailer/app"
)

type handler struct{}

func NewHandler(*app.Service) api.StrictServerInterface { return &handler{} }
