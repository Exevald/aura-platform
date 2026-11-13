package public

import (
	"context"
	"errors"
	api "filestorage/api/server/publicapi"
	"filestorage/internal/filestorage/app"
	"filestorage/internal/filestorage/domain"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type handler struct{ svc *app.Service }

func NewHandler(s *app.Service) api.StrictServerInterface { return &handler{s} }

func (h *handler) ListFiles(ctx context.Context, r api.ListFilesRequestObject) (api.ListFilesResponseObject, error) {
	v, e := h.svc.List(ctx, uuid.UUID(r.Params.XWorkspaceID))
	if e != nil {
		return api.ListFiles503JSONResponse{UnavailableJSONResponse: api.UnavailableJSONResponse{Message: "filestorage unavailable"}}, nil
	}
	out := make(api.ListFiles200JSONResponse, 0, len(v))
	for _, x := range v {
		out = append(out, toAPI(x))
	}
	return out, nil
}

func (h *handler) DeleteFile(ctx context.Context, r api.DeleteFileRequestObject) (api.DeleteFileResponseObject, error) {
	e := h.svc.Delete(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.FileID))
	if errors.Is(e, domain.ErrNotFound) || errors.Is(e, domain.ErrForbidden) {
		return api.DeleteFile404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse{Message: "file not found"}}, nil
	}
	if e != nil {
		return api.DeleteFile503JSONResponse{UnavailableJSONResponse: api.UnavailableJSONResponse{Message: "filestorage unavailable"}}, nil
	}
	return api.DeleteFile204Response{}, nil
}

func toAPI(v domain.File) api.File {
	return api.File{Id: openapi_types.UUID(v.ID), WorkspaceID: openapi_types.UUID(v.WorkspaceID), OwnerUserID: openapi_types.UUID(v.OwnerUserID), FileName: v.FileName, ContentType: v.ContentType, SizeBytes: v.SizeBytes, ChecksumSHA256: v.ChecksumSHA256, CreatedAt: v.CreatedAt}
}
