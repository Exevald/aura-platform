package public

import (
	"context"
	"errors"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"

	publicapi "shortlink/api/server/publicapi"
	"shortlink/internal/shortlink/app"
	"shortlink/internal/shortlink/domain"
)

type handler struct{ svc *app.Service }

func NewHandler(svc *app.Service) publicapi.StrictServerInterface { return &handler{svc: svc} }

func (h *handler) ListLinks(ctx context.Context, req publicapi.ListLinksRequestObject) (publicapi.ListLinksResponseObject, error) {
	links, err := h.svc.List(ctx, uuid.UUID(req.Params.XAuthUser))
	if err != nil {
		return publicapi.ListLinks503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(publicapi.ListLinks200JSONResponse, 0, len(links))
	for _, link := range links {
		out = append(out, toAPI(link))
	}
	return out, nil
}

func (h *handler) CreateLink(ctx context.Context, req publicapi.CreateLinkRequestObject) (publicapi.CreateLinkResponseObject, error) {
	link, err := h.svc.Create(ctx, uuid.UUID(req.Params.XAuthUser), req.Body.TargetURL)
	if errors.Is(err, domain.ErrInvalidTarget) {
		return publicapi.CreateLink400JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.CreateLink503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.CreateLink201JSONResponse(toAPI(link)), nil
}

func (h *handler) DeleteLink(ctx context.Context, req publicapi.DeleteLinkRequestObject) (publicapi.DeleteLinkResponseObject, error) {
	err := h.svc.Delete(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.LinkID))
	if errors.Is(err, domain.ErrNotFound) {
		return publicapi.DeleteLink404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.DeleteLink503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.DeleteLink204Response{}, nil
}

func (h *handler) ResolveLink(ctx context.Context, req publicapi.ResolveLinkRequestObject) (publicapi.ResolveLinkResponseObject, error) {
	link, err := h.svc.Resolve(ctx, req.Code)
	if errors.Is(err, domain.ErrNotFound) {
		return publicapi.ResolveLink404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.ResolveLink503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.ResolveLink302Response{Headers: publicapi.ResolveLink302ResponseHeaders{Location: link.TargetURL}}, nil
}

func toAPI(link domain.Link) publicapi.Link {
	return publicapi.Link{Id: openapi_types.UUID(link.ID), Code: link.Code, TargetURL: link.TargetURL, ShortPath: "/s/" + link.Code, CreatedAt: link.CreatedAt}
}

func unavailable() publicapi.UnavailableJSONResponse {
	return publicapi.UnavailableJSONResponse{Message: "shortlink unavailable"}
}
