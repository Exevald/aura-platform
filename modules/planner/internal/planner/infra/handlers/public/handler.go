package public

import (
	"context"
	"errors"
	api "planner/api/server/publicapi"
	"planner/internal/planner/app"
	"planner/internal/planner/domain"
	"uuid"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type handler struct{ svc *app.Service }

func NewHandler(s *app.Service) api.StrictServerInterface { return &handler{s} }

func (h *handler) ListProjects(ctx context.Context, r api.ListProjectsRequestObject) (api.ListProjectsResponseObject, error) {
	v, e := h.svc.ListProjects(ctx, uuid.UUID(r.Params.XWorkspaceID))
	if e != nil {
		return api.ListProjects503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(api.ListProjects200JSONResponse, 0, len(v))
	for _, x := range v {
		out = append(out, projectAPI(x))
	}
	return out, nil
}

func (h *handler) CreateProject(ctx context.Context, r api.CreateProjectRequestObject) (api.CreateProjectResponseObject, error) {
	v, e := h.svc.CreateProject(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), r.Body.Name)
	if errors.Is(e, domain.ErrInvalid) {
		return api.CreateProject400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if e != nil {
		return api.CreateProject503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.CreateProject201JSONResponse(projectAPI(v)), nil
}

func (h *handler) UpdateProject(ctx context.Context, r api.UpdateProjectRequestObject) (api.UpdateProjectResponseObject, error) {
	v, e := h.svc.UpdateProject(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.ProjectID), r.Body.Name)
	if errors.Is(e, domain.ErrInvalid) {
		return api.UpdateProject400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if errors.Is(e, domain.ErrNotFound) {
		return api.UpdateProject404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.UpdateProject503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.UpdateProject200JSONResponse(projectAPI(v)), nil
}

func (h *handler) DeleteProject(ctx context.Context, r api.DeleteProjectRequestObject) (api.DeleteProjectResponseObject, error) {
	e := h.svc.DeleteProject(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.ProjectID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.DeleteProject404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.DeleteProject503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.DeleteProject204Response{}, nil
}

func (h *handler) ListTasks(ctx context.Context, r api.ListTasksRequestObject) (api.ListTasksResponseObject, error) {
	v, e := h.svc.ListTasks(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.ProjectID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.ListTasks404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.ListTasks503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	out := make(api.ListTasks200JSONResponse, 0, len(v))
	for _, x := range v {
		out = append(out, taskAPI(x))
	}
	return out, nil
}

func (h *handler) CreateTask(ctx context.Context, r api.CreateTaskRequestObject) (api.CreateTaskResponseObject, error) {
	v, e := h.svc.CreateTask(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.ProjectID), taskInput(*r.Body, 0))
	if errors.Is(e, domain.ErrInvalid) {
		return api.CreateTask400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if errors.Is(e, domain.ErrNotFound) {
		return api.CreateTask404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.CreateTask503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.CreateTask201JSONResponse(taskAPI(v)), nil
}

func (h *handler) UpdateTask(ctx context.Context, r api.UpdateTaskRequestObject) (api.UpdateTaskResponseObject, error) {
	v, e := h.svc.UpdateTask(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.TaskID), taskUpdate(*r.Body))
	if errors.Is(e, domain.ErrInvalid) {
		return api.UpdateTask400JSONResponse{BadRequestJSONResponse: bad(e)}, nil
	}
	if errors.Is(e, domain.ErrNotFound) {
		return api.UpdateTask404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if errors.Is(e, domain.ErrConflict) {
		return api.UpdateTask409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse{Message: e.Error()}}, nil
	}
	if e != nil {
		return api.UpdateTask503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.UpdateTask200JSONResponse(taskAPI(v)), nil
}

func (h *handler) DeleteTask(ctx context.Context, r api.DeleteTaskRequestObject) (api.DeleteTaskResponseObject, error) {
	_, e := h.svc.DeleteTask(ctx, uuid.UUID(r.Params.XWorkspaceID), uuid.UUID(r.Params.XAuthUser), uuid.UUID(r.TaskID))
	if errors.Is(e, domain.ErrNotFound) {
		return api.DeleteTask404JSONResponse{NotFoundJSONResponse: notFound()}, nil
	}
	if e != nil {
		return api.DeleteTask503JSONResponse{UnavailableJSONResponse: unavailable()}, nil
	}
	return api.DeleteTask204Response{}, nil
}

func projectAPI(v domain.Project) api.Project {
	return api.Project{Id: openapi_types.UUID(v.ID), WorkspaceID: openapi_types.UUID(v.WorkspaceID), Name: v.Name, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func taskInput(v api.TaskInput, version int64) domain.Task {
	out := domain.Task{Title: v.Title, Description: v.Description, Status: string(v.Status), Priority: string(v.Priority), DueAt: v.DueAt, SyncToCalendar: v.SyncToCalendar, Version: version}
	if v.AssigneeUserID != nil {
		out.AssigneeUserID = uuid.UUID(*v.AssigneeUserID)
	}
	if v.CalendarID != nil {
		out.CalendarID = uuid.UUID(*v.CalendarID)
	}
	if v.Timezone != nil {
		out.Timezone = *v.Timezone
	}
	return out
}

func taskUpdate(v api.TaskUpdate) domain.Task {
	return taskInput(api.TaskInput{Title: v.Title, Description: v.Description, Status: api.TaskInputStatus(v.Status), Priority: api.TaskInputPriority(v.Priority), AssigneeUserID: v.AssigneeUserID, DueAt: v.DueAt, SyncToCalendar: v.SyncToCalendar, CalendarID: v.CalendarID, Timezone: v.Timezone}, v.Version)
}

func taskAPI(v domain.Task) api.Task {
	out := api.Task{Id: openapi_types.UUID(v.ID), ProjectID: openapi_types.UUID(v.ProjectID), Title: v.Title, Description: v.Description, Status: api.TaskStatus(v.Status), Priority: api.TaskPriority(v.Priority), SyncToCalendar: v.SyncToCalendar, Version: v.Version, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, DueAt: v.DueAt}
	if v.AssigneeUserID != uuid.Nil() {
		x := openapi_types.UUID(v.AssigneeUserID)
		out.AssigneeUserID = &x
	}
	if v.CalendarID != uuid.Nil() {
		x := openapi_types.UUID(v.CalendarID)
		out.CalendarID = &x
	}
	if v.CalendarEventID != uuid.Nil() {
		x := openapi_types.UUID(v.CalendarEventID)
		out.CalendarEventID = &x
	}
	if v.Timezone != "" {
		out.Timezone = &v.Timezone
	}
	return out
}

func bad(e error) api.BadRequestJSONResponse { return api.BadRequestJSONResponse{Message: e.Error()} }

func notFound() api.NotFoundJSONResponse {
	return api.NotFoundJSONResponse{Message: "planner resource not found"}
}

func unavailable() api.UnavailableJSONResponse {
	return api.UnavailableJSONResponse{Message: "planner unavailable"}
}
