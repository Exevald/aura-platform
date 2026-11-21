package public

import (
	"context"
	"errors"
	"publication/api/server/publicapi"
	"publication/internal/publication/app/model"
	"publication/internal/publication/app/service"
	"uuid"
)

type releaseAPI struct{ releases *service.Releases }

var _ publicapi.StrictServerInterface = (*releaseAPI)(nil)

func apiRelease(r model.Release) publicapi.Release {
	return publicapi.Release{Id: r.ID.String(), DocumentID: r.DocumentID.String(), WorkspaceID: r.WorkspaceID.String(), Status: r.Status, Title: r.Title, ShortPath: r.ShortPath, Failure: r.Failure, SourceVersion: r.SourceVersion, DeliveryStatus: r.DeliveryStatus}
}

func (h *releaseAPI) CreateRelease(ctx context.Context, req publicapi.CreateReleaseRequestObject) (publicapi.CreateReleaseResponseObject, error) {
	if req.Body == nil {
		return publicapi.CreateRelease400Response{}, nil
	}
	r, err := h.releases.Create(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.Body.DocumentID), req.Params.IdempotencyKey, req.Body.ExpectedVersion)
	switch {
	case errors.Is(err, service.ErrInvalidRequest):
		return publicapi.CreateRelease400Response{}, nil
	case errors.Is(err, service.ErrIdempotencyConflict):
		return publicapi.CreateRelease409Response{}, nil
	case err != nil:
		return publicapi.CreateRelease503Response{}, nil
	default:
		return publicapi.CreateRelease202JSONResponse(apiRelease(r)), nil
	}
}

func (h *releaseAPI) ListReleases(ctx context.Context, req publicapi.ListReleasesRequestObject) (publicapi.ListReleasesResponseObject, error) {
	rows, err := h.releases.List(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID))
	if err != nil {
		return publicapi.ListReleases503Response{}, nil
	}
	out := make(publicapi.ListReleases200JSONResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, apiRelease(r))
	}
	return out, nil
}

func (h *releaseAPI) GetRelease(ctx context.Context, req publicapi.GetReleaseRequestObject) (publicapi.GetReleaseResponseObject, error) {
	r, err := h.releases.Get(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.Id))
	if errors.Is(err, service.ErrNotFound) {
		return publicapi.GetRelease404Response{}, nil
	}
	if err != nil {
		return publicapi.GetRelease503Response{}, nil
	}
	return publicapi.GetRelease200JSONResponse(apiRelease(r)), nil
}

func (h *releaseAPI) CloseRelease(ctx context.Context, req publicapi.CloseReleaseRequestObject) (publicapi.CloseReleaseResponseObject, error) {
	err := h.releases.Close(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.Id))
	if errors.Is(err, service.ErrNotFound) {
		return publicapi.CloseRelease404Response{}, nil
	}
	if err != nil {
		return publicapi.CloseRelease503Response{}, nil
	}
	return publicapi.CloseRelease204Response{}, nil
}
