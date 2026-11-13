package public

import (
	"context"
	"errors"
	"uuid"
	publicapi "workspace/api/server/publicapi"
	"workspace/internal/workspace/app"
	"workspace/internal/workspace/domain"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const workspaceNotFoundMessage = "workspace not found"

type handler struct{ svc *app.Service }

func NewHandler(s *app.Service) publicapi.StrictServerInterface { return &handler{svc: s} }

func (h *handler) ListWorkspaces(ctx context.Context, r publicapi.ListWorkspacesRequestObject) (publicapi.ListWorkspacesResponseObject, error) {
	items, e := h.svc.List(ctx, uuid.UUID(r.Params.XAuthUser))
	if e != nil {
		return publicapi.ListWorkspaces503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(publicapi.ListWorkspaces200JSONResponse, 0, len(items))
	for _, w := range items {
		_, m, role, e := h.svc.Get(ctx, uuid.UUID(r.Params.XAuthUser), w.ID)
		if e != nil {
			return publicapi.ListWorkspaces503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
		}
		out = append(out, toAPI(w, m, role))
	}
	return out, nil
}

func (h *handler) CreateWorkspace(ctx context.Context, r publicapi.CreateWorkspaceRequestObject) (publicapi.CreateWorkspaceResponseObject, error) {
	w, e := h.svc.Create(ctx, uuid.UUID(r.Params.XAuthUser), r.Body.Name)
	if errors.Is(e, domain.ErrInvalid) {
		return publicapi.CreateWorkspace400JSONResponse{BadRequestJSONResponse: publicapi.BadRequestJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return publicapi.CreateWorkspace503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	_, m, role, e := h.svc.Get(ctx, uuid.UUID(r.Params.XAuthUser), w.ID)
	if e != nil {
		return publicapi.CreateWorkspace503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.CreateWorkspace201JSONResponse(toAPI(w, m, role)), nil
}

func (h *handler) GetWorkspace(ctx context.Context, r publicapi.GetWorkspaceRequestObject) (publicapi.GetWorkspaceResponseObject, error) {
	w, m, role, e := h.svc.Get(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID))
	if errors.Is(e, domain.ErrNotFound) {
		return publicapi.GetWorkspace404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return publicapi.GetWorkspace503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.GetWorkspace200JSONResponse(toAPI(w, m, role)), nil
}

func (h *handler) DeleteWorkspace(ctx context.Context, r publicapi.DeleteWorkspaceRequestObject) (publicapi.DeleteWorkspaceResponseObject, error) {
	e := h.svc.Delete(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID))
	if errors.Is(e, domain.ErrNotFound) || errors.Is(e, domain.ErrForbidden) {
		return publicapi.DeleteWorkspace404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: workspaceNotFoundMessage}}, nil
	}
	if errors.Is(e, domain.ErrPersonal) {
		return publicapi.DeleteWorkspace409JSONResponse{ConflictJSONResponse: publicapi.ConflictJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return publicapi.DeleteWorkspace503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.DeleteWorkspace204Response{}, nil
}

func (h *handler) PutWorkspaceMember(ctx context.Context, r publicapi.PutWorkspaceMemberRequestObject) (publicapi.PutWorkspaceMemberResponseObject, error) {
	e := h.svc.PutMember(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID), uuid.UUID(r.Body.UserID), domain.Role(r.Body.Role))
	if errors.Is(e, domain.ErrInvalid) {
		return publicapi.PutWorkspaceMember400JSONResponse{BadRequestJSONResponse: publicapi.BadRequestJSONResponse{Message: e.Error()}}, nil
	}
	if errors.Is(e, domain.ErrNotFound) || errors.Is(e, domain.ErrForbidden) {
		return publicapi.PutWorkspaceMember404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: workspaceNotFoundMessage}}, nil
	}
	if e != nil {
		return publicapi.PutWorkspaceMember503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	w, m, role, e := h.svc.Get(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID))
	if e != nil {
		return publicapi.PutWorkspaceMember503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.PutWorkspaceMember200JSONResponse(toAPI(w, m, role)), nil
}

func (h *handler) RemoveWorkspaceMember(ctx context.Context, r publicapi.RemoveWorkspaceMemberRequestObject) (publicapi.RemoveWorkspaceMemberResponseObject, error) {
	e := h.svc.RemoveMember(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID), uuid.UUID(r.UserID))
	if errors.Is(e, domain.ErrInvalid) {
		return publicapi.RemoveWorkspaceMember409JSONResponse{ConflictJSONResponse: publicapi.ConflictJSONResponse{Message: e.Error()}}, nil
	}
	if errors.Is(e, domain.ErrNotFound) || errors.Is(e, domain.ErrForbidden) {
		return publicapi.RemoveWorkspaceMember404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: workspaceNotFoundMessage}}, nil
	}
	if e != nil {
		return publicapi.RemoveWorkspaceMember503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.RemoveWorkspaceMember204Response{}, nil
}

func (h *handler) CreateWorkspaceInvitation(ctx context.Context, r publicapi.CreateWorkspaceInvitationRequestObject) (publicapi.CreateWorkspaceInvitationResponseObject, error) {
	user := uuid.Nil()
	if r.Body.UserID != nil {
		user = uuid.UUID(*r.Body.UserID)
	}
	v, e := h.svc.CreateInvitation(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID), user, string(r.Body.Email), domain.Role(r.Body.Role))
	if errors.Is(e, domain.ErrInvalid) {
		return publicapi.CreateWorkspaceInvitation400JSONResponse{BadRequestJSONResponse: publicapi.BadRequestJSONResponse{Message: e.Error()}}, nil
	}
	if errors.Is(e, domain.ErrNotFound) || errors.Is(e, domain.ErrForbidden) {
		return publicapi.CreateWorkspaceInvitation404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: workspaceNotFoundMessage}}, nil
	}
	if e != nil {
		return publicapi.CreateWorkspaceInvitation503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.CreateWorkspaceInvitation201JSONResponse(invitationToAPI(v)), nil
}

func (h *handler) ListWorkspaceInvitations(ctx context.Context, r publicapi.ListWorkspaceInvitationsRequestObject) (publicapi.ListWorkspaceInvitationsResponseObject, error) {
	items, e := h.svc.ListInvitations(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID))
	if errors.Is(e, domain.ErrNotFound) {
		return publicapi.ListWorkspaceInvitations404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: workspaceNotFoundMessage}}, nil
	}
	if e != nil {
		return publicapi.ListWorkspaceInvitations503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(publicapi.ListWorkspaceInvitations200JSONResponse, 0, len(items))
	for _, v := range items {
		out = append(out, invitationToAPI(v))
	}
	return out, nil
}

func (h *handler) AcceptWorkspaceInvitation(ctx context.Context, r publicapi.AcceptWorkspaceInvitationRequestObject) (publicapi.AcceptWorkspaceInvitationResponseObject, error) {
	actor := uuid.UUID(r.Params.XAuthUser)
	w, e := h.svc.AcceptInvitation(ctx, actor, r.Token)
	if errors.Is(e, domain.ErrNotFound) {
		return publicapi.AcceptWorkspaceInvitation404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: "invitation not found"}}, nil
	}
	if errors.Is(e, domain.ErrExpired) {
		return publicapi.AcceptWorkspaceInvitation409JSONResponse{ConflictJSONResponse: publicapi.ConflictJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return publicapi.AcceptWorkspaceInvitation503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	_, members, role, e := h.svc.Get(ctx, actor, w.ID)
	if e != nil {
		return publicapi.AcceptWorkspaceInvitation503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.AcceptWorkspaceInvitation200JSONResponse(toAPI(w, members, role)), nil
}

func (h *handler) RevokeWorkspaceInvitation(ctx context.Context, r publicapi.RevokeWorkspaceInvitationRequestObject) (publicapi.RevokeWorkspaceInvitationResponseObject, error) {
	e := h.svc.RevokeInvitation(ctx, uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.WorkspaceID), uuid.UUID(r.InvitationID))
	if errors.Is(e, domain.ErrNotFound) {
		return publicapi.RevokeWorkspaceInvitation404JSONResponse{NotFoundJSONResponse: publicapi.NotFoundJSONResponse{Message: "invitation not found"}}, nil
	}
	if e != nil {
		return publicapi.RevokeWorkspaceInvitation503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return publicapi.RevokeWorkspaceInvitation204Response{}, nil
}

func invitationToAPI(v domain.Invitation) publicapi.Invitation {
	out := publicapi.Invitation{Id: openapi_types.UUID(v.ID), WorkspaceID: openapi_types.UUID(v.WorkspaceID), Email: openapi_types.Email(v.Email), Role: publicapi.InvitationRole(v.Role), Token: v.Token, ExpiresAt: v.ExpiresAt, CreatedAt: v.CreatedAt}
	if v.UserID != uuid.Nil() {
		id := openapi_types.UUID(v.UserID)
		out.UserID = &id
	}
	return out
}

func toAPI(w domain.Workspace, members []domain.Member, role domain.Role) publicapi.Workspace {
	m := make([]publicapi.Member, 0, len(members))
	for _, v := range members {
		m = append(m, publicapi.Member{UserID: openapi_types.UUID(v.UserID), Role: publicapi.Role(v.Role), CreatedAt: v.CreatedAt})
	}
	return publicapi.Workspace{Id: openapi_types.UUID(w.ID), Kind: publicapi.WorkspaceKind(w.Kind), Name: w.Name, Role: publicapi.Role(role), Members: m, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt}
}

func unavailable() publicapi.UnavailableJSONResponse {
	return publicapi.UnavailableJSONResponse{Message: "workspace unavailable"}
}
