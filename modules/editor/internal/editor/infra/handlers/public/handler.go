package public

import (
	"context"
	"errors"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	publicapi "editor/api/server/publicapi"
	"editor/internal/editor/app"
	"editor/internal/editor/domain"
)

const (
	editorUnavailable      = "editor unavailable"
	permissionsUnavailable = "permissions unavailable"
)

func NewHandler(svc app.EditorService) publicapi.StrictServerInterface {
	return &handler{svc: svc}
}

type handler struct {
	svc app.EditorService
}

func (h *handler) ListDocuments(
	ctx context.Context,
	req publicapi.ListDocumentsRequestObject,
) (publicapi.ListDocumentsResponseObject, error) {
	values, err := h.svc.List(ctx, convertToUserID(req.Params.XAuthUser))
	if err != nil {
		return publicapi.ListDocuments503JSONResponse{Message: editorUnavailable}, nil
	}
	out := make(publicapi.ListDocuments200JSONResponse, 0, len(values))
	for _, v := range values {
		out = append(out, convertToAPIDocument(v))
	}
	return out, nil
}

func (h *handler) CreateDocument(
	ctx context.Context,
	req publicapi.CreateDocumentRequestObject,
) (publicapi.CreateDocumentResponseObject, error) {
	d, err := h.svc.Create(ctx, convertToUserID(req.Params.XAuthUser), req.Body.Title, req.Body.Content)
	if errors.Is(err, domain.ErrInvalidDocument) {
		return publicapi.CreateDocument400JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.CreateDocument503JSONResponse{Message: editorUnavailable}, nil
	}
	return publicapi.CreateDocument201JSONResponse(convertToAPIDocument(d)), nil
}

func (h *handler) GetDocument(
	ctx context.Context,
	req publicapi.GetDocumentRequestObject,
) (publicapi.GetDocumentResponseObject, error) {
	d, err := h.svc.Get(ctx, convertToUserID(req.Params.XAuthUser), uuid.UUID(req.DocumentID))
	if errors.Is(err, domain.ErrPermissionDenied) {
		return publicapi.GetDocument403JSONResponse{Message: err.Error()}, nil
	}
	if errors.Is(err, domain.ErrDocumentNotFound) {
		return publicapi.GetDocument404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.GetDocument503JSONResponse{Message: editorUnavailable}, nil
	}
	return publicapi.GetDocument200JSONResponse(convertToAPIDocument(d)), nil
}

func (h *handler) UpdateDocument(
	ctx context.Context,
	req publicapi.UpdateDocumentRequestObject,
) (publicapi.UpdateDocumentResponseObject, error) {
	d, err := h.svc.Update(
		ctx,
		convertToUserID(req.Params.XAuthUser),
		uuid.UUID(req.DocumentID),
		req.Body.Title,
		req.Body.Content,
		req.Body.Version,
	)
	switch {
	case errors.Is(err, domain.ErrInvalidDocument):
		return publicapi.UpdateDocument400JSONResponse{
			BadRequestJSONResponse: publicapi.BadRequestJSONResponse{Message: err.Error()},
		}, nil
	case errors.Is(err, domain.ErrPermissionDenied):
		return publicapi.UpdateDocument403JSONResponse{
			ForbiddenJSONResponse: publicapi.ForbiddenJSONResponse{Message: err.Error()},
		}, nil
	case errors.Is(err, domain.ErrDocumentNotFound):
		return publicapi.UpdateDocument404JSONResponse{
			NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: err.Error()},
		}, nil
	case errors.Is(err, domain.ErrVersionConflict):
		return publicapi.UpdateDocument409JSONResponse{
			ConflictJSONResponse: publicapi.ConflictJSONResponse{Message: err.Error()},
		}, nil
	case err != nil:
		return publicapi.UpdateDocument503JSONResponse{
			UnavailableJSONResponse: publicapi.UnavailableJSONResponse{Message: editorUnavailable},
		}, nil
	}
	return publicapi.UpdateDocument200JSONResponse(convertToAPIDocument(d)), nil
}

func (h *handler) DeleteDocument(
	ctx context.Context,
	req publicapi.DeleteDocumentRequestObject,
) (publicapi.DeleteDocumentResponseObject, error) {
	err := h.svc.Delete(
		ctx,
		convertToUserID(req.Params.XAuthUser),
		uuid.UUID(req.DocumentID),
	)
	if errors.Is(err, domain.ErrPermissionDenied) {
		return publicapi.DeleteDocument403JSONResponse{
			ForbiddenJSONResponse: publicapi.ForbiddenJSONResponse{Message: err.Error()},
		}, nil
	}
	if errors.Is(err, domain.ErrDocumentNotFound) {
		return publicapi.DeleteDocument404JSONResponse{
			NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: err.Error()},
		}, nil
	}
	if err != nil {
		return publicapi.DeleteDocument503JSONResponse{
			UnavailableJSONResponse: publicapi.UnavailableJSONResponse{Message: editorUnavailable},
		}, nil
	}
	return publicapi.DeleteDocument204Response{}, nil
}

func (h *handler) ListPermissions(
	ctx context.Context,
	req publicapi.ListPermissionsRequestObject,
) (publicapi.ListPermissionsResponseObject, error) {
	values, err := h.svc.Permissions(
		ctx,
		convertToUserID(req.Params.XAuthUser),
		uuid.UUID(req.DocumentID),
	)
	if errors.Is(err, domain.ErrPermissionDenied) {
		return publicapi.ListPermissions403JSONResponse{
			ForbiddenJSONResponse: publicapi.ForbiddenJSONResponse{Message: err.Error()},
		}, nil
	}
	if err != nil {
		return publicapi.ListPermissions503JSONResponse{
			UnavailableJSONResponse: publicapi.UnavailableJSONResponse{Message: permissionsUnavailable},
		}, nil
	}
	resultPermissions := make(publicapi.ListPermissions200JSONResponse, 0, len(values))
	for _, v := range values {
		resultPermissions = append(
			resultPermissions,
			publicapi.Permission{
				UserID:      openapi_types.UUID(v.UserID),
				Email:       openapi_types.Email(v.Email),
				DisplayName: v.DisplayName,
				Role:        publicapi.PermissionRole(v.Role),
				UpdatedAt:   v.UpdatedAt,
			})
	}
	return resultPermissions, nil
}

func (h *handler) GrantPermission(
	ctx context.Context,
	req publicapi.GrantPermissionRequestObject,
) (publicapi.GrantPermissionResponseObject, error) {
	err := h.svc.Grant(
		ctx,
		convertToUserID(req.Params.XAuthUser),
		uuid.UUID(req.DocumentID),
		string(req.Body.Email),
		domain.Role(req.Body.Role),
	)
	if errors.Is(err, domain.ErrInvalidPermission) {
		return publicapi.GrantPermission400JSONResponse{
			BadRequestJSONResponse: publicapi.BadRequestJSONResponse{Message: err.Error()},
		}, nil
	}
	if errors.Is(err, domain.ErrPermissionDenied) {
		return publicapi.GrantPermission403JSONResponse{
			ForbiddenJSONResponse: publicapi.ForbiddenJSONResponse{Message: err.Error()},
		}, nil
	}
	if status.Code(err) == codes.NotFound {
		return publicapi.GrantPermission404JSONResponse{
			NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: "user not found"},
		}, nil
	}
	if err != nil {
		return publicapi.GrantPermission503JSONResponse{
			UnavailableJSONResponse: publicapi.UnavailableJSONResponse{Message: permissionsUnavailable},
		}, nil
	}
	return publicapi.GrantPermission204Response{}, nil
}

func (h *handler) RevokePermission(
	ctx context.Context,
	req publicapi.RevokePermissionRequestObject,
) (publicapi.RevokePermissionResponseObject, error) {
	err := h.svc.Revoke(
		ctx,
		convertToUserID(req.Params.XAuthUser),
		uuid.UUID(req.DocumentID),
		uuid.UUID(req.UserID),
	)
	if errors.Is(err, domain.ErrInvalidPermission) {
		return publicapi.RevokePermission400JSONResponse{
			BadRequestJSONResponse: publicapi.BadRequestJSONResponse{Message: err.Error()},
		}, nil
	}
	if errors.Is(err, domain.ErrPermissionDenied) {
		return publicapi.RevokePermission403JSONResponse{
			ForbiddenJSONResponse: publicapi.ForbiddenJSONResponse{Message: err.Error()},
		}, nil
	}
	if err != nil {
		return publicapi.RevokePermission503JSONResponse{
			UnavailableJSONResponse: publicapi.UnavailableJSONResponse{Message: permissionsUnavailable},
		}, nil
	}
	return publicapi.RevokePermission204Response{}, nil
}

func convertToUserID(id openapi_types.UUID) uuid.UUID {
	return uuid.UUID(id)
}

func convertToAPIDocument(d domain.Document) publicapi.Document {
	return publicapi.Document{
		Id:          openapi_types.UUID(d.ID),
		OwnerUserID: openapi_types.UUID(d.OwnerUserID),
		Title:       d.Title,
		Content:     d.Content,
		Version:     d.Version,
		Role:        publicapi.DocumentRole(d.Role),
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
