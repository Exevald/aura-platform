package public

import (
	"context"
	"errors"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	publicapi "usergroup/api/server/publicapi"
	"usergroup/internal/usergroup/app"
	"usergroup/internal/usergroup/domain"
)

type handler struct{ svc *app.Service }

func NewHandler(svc *app.Service) publicapi.StrictServerInterface { return &handler{svc: svc} }

func (h *handler) ListGroups(ctx context.Context, req publicapi.ListGroupsRequestObject) (publicapi.ListGroupsResponseObject, error) {
	values, err := h.svc.List(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID))
	if err != nil {
		return publicapi.ListGroups503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(publicapi.ListGroups200JSONResponse, 0, len(values))
	for _, value := range values {
		out = append(out, toAPI(value))
	}
	return out, nil
}

func (h *handler) CreateGroup(ctx context.Context, req publicapi.CreateGroupRequestObject) (publicapi.CreateGroupResponseObject, error) {
	value, err := h.svc.Create(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), req.Body.Name)
	if errors.Is(err, domain.ErrInvalidGroup) {
		return publicapi.CreateGroup400JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.CreateGroup503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.CreateGroup201JSONResponse(toAPI(value)), nil
}

func (h *handler) GetGroup(ctx context.Context, req publicapi.GetGroupRequestObject) (publicapi.GetGroupResponseObject, error) {
	value, err := h.svc.Get(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.GroupID))
	if errors.Is(err, domain.ErrGroupNotFound) {
		return publicapi.GetGroup404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.GetGroup503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.GetGroup200JSONResponse(toAPI(value)), nil
}

func (h *handler) UpdateGroup(ctx context.Context, req publicapi.UpdateGroupRequestObject) (publicapi.UpdateGroupResponseObject, error) {
	value, err := h.svc.Update(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.GroupID), req.Body.Name)
	if errors.Is(err, domain.ErrInvalidGroup) {
		return publicapi.UpdateGroup400JSONResponse{Message: err.Error()}, nil
	}
	if errors.Is(err, domain.ErrGroupNotFound) {
		return publicapi.UpdateGroup404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.UpdateGroup503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.UpdateGroup200JSONResponse(toAPI(value)), nil
}

func (h *handler) DeleteGroup(ctx context.Context, req publicapi.DeleteGroupRequestObject) (publicapi.DeleteGroupResponseObject, error) {
	err := h.svc.Delete(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.GroupID))
	if errors.Is(err, domain.ErrGroupNotFound) {
		return publicapi.DeleteGroup404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.DeleteGroup503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.DeleteGroup204Response{}, nil
}

func (h *handler) AddGroupMember(ctx context.Context, req publicapi.AddGroupMemberRequestObject) (publicapi.AddGroupMemberResponseObject, error) {
	value, err := h.svc.AddMember(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.GroupID), string(req.Body.Email))
	if status.Code(err) == codes.NotFound || errors.Is(err, domain.ErrGroupNotFound) {
		return publicapi.AddGroupMember404JSONResponse{Message: "group or user not found"}, nil
	}
	if errors.Is(err, domain.ErrInvalidGroup) {
		return publicapi.AddGroupMember400JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.AddGroupMember503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.AddGroupMember200JSONResponse(toAPI(value)), nil
}

func (h *handler) RemoveGroupMember(ctx context.Context, req publicapi.RemoveGroupMemberRequestObject) (publicapi.RemoveGroupMemberResponseObject, error) {
	err := h.svc.RemoveMember(ctx, uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.Params.XWorkspaceID), uuid.UUID(req.GroupID), uuid.UUID(req.UserID))
	if errors.Is(err, domain.ErrGroupNotFound) {
		return publicapi.RemoveGroupMember404JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.RemoveGroupMember503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.RemoveGroupMember204Response{}, nil
}

func toAPI(value app.Group) publicapi.Group {
	members := make([]publicapi.Member, 0, len(value.Members))
	for _, member := range value.Members {
		members = append(members, publicapi.Member{UserID: openapi_types.UUID(member.UserID), Email: openapi_types.Email(member.Email), DisplayName: member.DisplayName})
	}
	return publicapi.Group{Id: openapi_types.UUID(value.ID), WorkspaceID: openapi_types.UUID(value.WorkspaceID), OwnerUserID: openapi_types.UUID(value.OwnerUserID), Name: value.Name, Members: members, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}
}

func unavailable() publicapi.UnavailableJSONResponse {
	return publicapi.UnavailableJSONResponse{Message: "groups unavailable"}
}
