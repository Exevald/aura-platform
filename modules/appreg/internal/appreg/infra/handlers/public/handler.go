package public

import (
	api "appreg/api/server/publicapi"
	"appreg/internal/appreg/app"
	"context"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type handler struct{ svc *app.Service }

func NewHandler(s *app.Service) api.StrictServerInterface { return &handler{svc: s} }

func (h *handler) ListApps(context.Context, api.ListAppsRequestObject) (api.ListAppsResponseObject, error) {
	values := h.svc.List(false)
	out := make(api.ListApps200JSONResponse, 0, len(values))
	for _, v := range values {
		out = append(out, toAPI(v))
	}
	return out, nil
}

func toAPI(v app.App) api.App {
	return api.App{Id: openapi_types.UUID(v.ID), Code: v.Code, Name: v.Name, FrontendRoute: v.FrontendRoute, BackendTarget: v.BackendTarget, DefaultEnabled: v.DefaultEnabled}
}
