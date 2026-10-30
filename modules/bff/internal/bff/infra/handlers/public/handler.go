package public

import (
	"context"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	publicapi "bff/api/server/publicapi"
	"bff/internal/bff/app"
)

func NewHandler(svc *app.Service) publicapi.StrictServerInterface { return &handler{svc: svc} }

type handler struct{ svc *app.Service }

func (h *handler) GetMe(
	ctx context.Context,
	req publicapi.GetMeRequestObject,
) (publicapi.GetMeResponseObject, error) {
	id, err := uuid.Parse(req.Params.XAuthUser.String())
	if err != nil {
		return publicapi.GetMe400JSONResponse{Message: "invalid user context"}, nil
	}
	p, err := h.svc.GetMe(ctx, id)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return publicapi.GetMe404JSONResponse{Message: "user not found"}, nil
		}
		return publicapi.GetMe503JSONResponse{Message: "profile service unavailable"}, nil
	}
	return getResponse(p), nil
}

func (h *handler) UpdateMe(
	ctx context.Context,
	req publicapi.UpdateMeRequestObject,
) (publicapi.UpdateMeResponseObject, error) {
	id, err := uuid.Parse(req.Params.XAuthUser.String())
	if err != nil {
		return publicapi.UpdateMe400JSONResponse{Message: "invalid user context"}, nil
	}
	p, err := h.svc.UpdateMe(
		ctx,
		app.Profile{
			UserID:      id,
			DisplayName: req.Body.DisplayName,
			FirstName:   req.Body.FirstName,
			LastName:    req.Body.LastName,
			Email:       string(req.Body.Email),
		})
	if err != nil {
		switch status.Code(err) {
		case codes.AlreadyExists:
			return publicapi.UpdateMe409JSONResponse{Message: "email is already in use"}, nil
		case codes.InvalidArgument:
			return publicapi.UpdateMe400JSONResponse{Message: "invalid profile"}, nil
		default:
			return publicapi.UpdateMe503JSONResponse{Message: "profile service unavailable"}, nil
		}
	}
	return updateResponse(p), nil
}

func getResponse(p app.Profile) publicapi.GetMe200JSONResponse {
	return publicapi.GetMe200JSONResponse{UserID: openapi_types.UUID(p.UserID), DisplayName: p.DisplayName, FirstName: p.FirstName, LastName: p.LastName, Email: openapi_types.Email(p.Email)}
}
func updateResponse(p app.Profile) publicapi.UpdateMe200JSONResponse {
	return publicapi.UpdateMe200JSONResponse{UserID: openapi_types.UUID(p.UserID), DisplayName: p.DisplayName, FirstName: p.FirstName, LastName: p.LastName, Email: openapi_types.Email(p.Email)}
}
