package public

import (
	"context"
	"errors"
	api "notification/api/server/publicapi"
	"notification/internal/notification/app"
	"notification/internal/notification/domain"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type handler struct{ svc *app.Service }

func NewHandler(s *app.Service) api.StrictServerInterface { return &handler{s} }

func (h *handler) ListNotifications(ctx context.Context, r api.ListNotificationsRequestObject) (api.ListNotificationsResponseObject, error) {
	values, e := h.svc.List(ctx, uuid.UUID(r.Params.XAuthUser))
	if e != nil {
		return api.ListNotifications503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(api.ListNotifications200JSONResponse, 0, len(values))
	for _, v := range values {
		out = append(out, toAPI(v))
	}
	return out, nil
}

func (h *handler) MarkNotificationRead(ctx context.Context, r api.MarkNotificationReadRequestObject) (api.MarkNotificationReadResponseObject, error) {
	e := h.svc.MarkRead(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.NotificationID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.MarkNotificationRead404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return api.MarkNotificationRead503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.MarkNotificationRead204Response{}, nil
}

func (h *handler) GetPreferences(ctx context.Context, r api.GetPreferencesRequestObject) (api.GetPreferencesResponseObject, error) {
	v, e := h.svc.Preferences(ctx, uuid.UUID(r.Params.XAuthUser))
	if e != nil {
		return nil, e
	}
	return api.GetPreferences200JSONResponse{InApp: v.InApp, Email: v.Email}, nil
}

func (h *handler) PutPreferences(ctx context.Context, r api.PutPreferencesRequestObject) (api.PutPreferencesResponseObject, error) {
	v, e := h.svc.PutPreferences(ctx, domain.Preferences{UserID: uuid.UUID(r.Params.XAuthUser), InApp: r.Body.InApp, Email: r.Body.Email})
	if e != nil {
		return api.PutPreferences503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.PutPreferences200JSONResponse{InApp: v.InApp, Email: v.Email}, nil
}

func toAPI(v domain.Notification) api.Notification {
	out := api.Notification{Id: openapi_types.UUID(v.ID), Kind: v.Kind, Title: v.Title, Body: v.Body, Read: v.Read, CreatedAt: v.CreatedAt}
	if v.WorkspaceID != uuid.Nil() {
		id := openapi_types.UUID(v.WorkspaceID)
		out.WorkspaceID = &id
	}
	return out
}

func unavailable() api.UnavailableJSONResponse {
	return api.UnavailableJSONResponse{Message: "notifications unavailable"}
}
