package public

import (
	"context"
	"errors"
	"uuid"

	openapitypes "github.com/oapi-codegen/runtime/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"access/api/server/publicapi"
	"access/internal/access/app"
	"access/internal/access/domain/model"
	clientinfra "access/internal/access/infra/client"
)

type handler struct {
	svc      app.ModuleAccessService
	groups   clientinfra.UserGroup
	registry clientinfra.AppRegistry
}

func NewHandler(svc app.ModuleAccessService, groups clientinfra.UserGroup, registries ...clientinfra.AppRegistry) publicapi.StrictServerInterface {
	h := &handler{svc: svc, groups: groups}
	if len(registries) > 0 {
		h.registry = registries[0]
	}
	return h
}

func (h *handler) ListModules(ctx context.Context, _ publicapi.ListModulesRequestObject) (publicapi.ListModulesResponseObject, error) {
	modules := app.Modules()
	if h.registry != nil {
		var err error
		modules, err = h.registry.List(ctx)
		if err != nil {
			return publicapi.ListModules503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
		}
	}
	out := make(publicapi.ListModules200JSONResponse, 0, len(modules))
	for _, module := range modules {
		out = append(out,
			publicapi.Module{
				Id:   openapitypes.UUID(module.ID),
				Code: module.Code,
				Name: module.Name,
			})
	}
	return out, nil
}

func (h *handler) ListGroupModuleAccesses(ctx context.Context, req publicapi.ListGroupModuleAccessesRequestObject) (publicapi.ListGroupModuleAccessesResponseObject, error) {
	values, err := h.svc.ListGroupModuleAccess(ctx, uuid.UUID(req.Params.XAuthUser))
	if err != nil {
		return publicapi.ListGroupModuleAccesses503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	grouped := groupAccesses(values)
	out := make(publicapi.ListGroupModuleAccesses200JSONResponse, 0, len(grouped))
	for groupID, moduleIDs := range grouped {
		out = append(out, toAPI(groupID, moduleIDs))
	}
	return out, nil
}

func (h *handler) SetGroupModules(ctx context.Context, req publicapi.SetGroupModulesRequestObject) (publicapi.SetGroupModulesResponseObject, error) {
	ownerID, groupID := uuid.UUID(req.Params.XAuthUser), uuid.UUID(req.GroupID)
	members, err := h.groups.Members(ctx, groupID, ownerID)
	if status.Code(err) == codes.NotFound {
		return publicapi.SetGroupModules404JSONResponse{Message: "group not found"}, nil
	}
	if err != nil {
		return publicapi.SetGroupModules503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	members = moduleRecipients(ownerID, members)
	moduleIDs := make([]uuid.UUID, 0, len(req.Body.ModuleIDs))
	for _, moduleID := range req.Body.ModuleIDs {
		moduleIDs = append(moduleIDs, uuid.UUID(moduleID))
	}
	values, err := h.svc.SetGroupModules(ctx, ownerID, groupID, moduleIDs, members)
	if errors.Is(err, model.ErrUnknownModule) {
		return publicapi.SetGroupModules400JSONResponse{Message: err.Error()}, nil
	}
	if errors.Is(err, model.ErrEmptyGroup) {
		return publicapi.SetGroupModules409JSONResponse{Message: err.Error()}, nil
	}
	if err != nil {
		return publicapi.SetGroupModules503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	ids := make([]uuid.UUID, 0, len(values))
	for _, value := range values {
		ids = append(ids, value.ModuleID)
	}
	return publicapi.SetGroupModules200JSONResponse(toAPI(groupID, ids)), nil
}

func moduleRecipients(ownerID uuid.UUID, members []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(members)+1)
	seen := make(map[uuid.UUID]struct{}, len(members)+1)
	for _, userID := range append([]uuid.UUID{ownerID}, members...) {
		if userID == uuid.Nil() {
			continue
		}
		if _, exists := seen[userID]; exists {
			continue
		}
		seen[userID] = struct{}{}
		result = append(result, userID)
	}
	return result
}

func groupAccesses(values []model.ModuleShare) map[uuid.UUID][]uuid.UUID {
	result := make(map[uuid.UUID][]uuid.UUID)
	seen := make(map[uuid.UUID]map[uuid.UUID]struct{})
	for _, value := range values {
		if seen[value.GroupID] == nil {
			seen[value.GroupID] = make(map[uuid.UUID]struct{})
		}
		if _, exists := seen[value.GroupID][value.ModuleID]; exists {
			continue
		}
		seen[value.GroupID][value.ModuleID] = struct{}{}
		result[value.GroupID] = append(result[value.GroupID], value.ModuleID)
	}
	return result
}

func toAPI(groupID uuid.UUID, moduleIDs []uuid.UUID) publicapi.GroupModuleAccess {
	ids := make([]openapitypes.UUID, 0, len(moduleIDs))
	for _, moduleID := range moduleIDs {
		ids = append(ids, openapitypes.UUID(moduleID))
	}
	return publicapi.GroupModuleAccess{GroupID: openapitypes.UUID(groupID), ModuleIDs: ids}
}

func unavailable() publicapi.UnavailableJSONResponse {
	return publicapi.UnavailableJSONResponse{Message: "access unavailable"}
}
