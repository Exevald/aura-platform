package internalapi

import (
	"context"
	"errors"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userinternalapi "user/api/server/internalapi"
	"user/internal/user/app"
	"user/internal/user/domain"
)

func NewHandler(svc app.UserService) userinternalapi.UserServiceServer {
	return &handler{svc: svc}
}

type handler struct {
	svc app.UserService
	userinternalapi.UnimplementedUserServiceServer
}

func (h *handler) CreateUser(
	ctx context.Context,
	req *userinternalapi.CreateUserRequest,
) (*userinternalapi.CreateUserResponse, error) {
	fields := req.GetFields()
	if fields == nil {
		return nil, status.Error(codes.InvalidArgument, "fields are required")
	}
	id, err := h.svc.CreateUser(ctx, fieldsSpec(fields))
	if err != nil {
		return nil, mapError(err)
	}
	return &userinternalapi.CreateUserResponse{UserId: id.String()}, nil
}

func (h *handler) GetUser(
	ctx context.Context,
	req *userinternalapi.GetUserRequest,
) (*userinternalapi.GetUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	user, err := h.svc.GetUser(ctx, id)
	if err != nil {
		return nil, mapError(err)
	}
	return &userinternalapi.GetUserResponse{
		UserId: user.ID.String(),
		Fields: &userinternalapi.UserFields{
			DisplayName: user.DisplayName,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
		},
	}, nil
}

func (h *handler) FindUserByEmail(
	ctx context.Context,
	req *userinternalapi.FindUserByEmailRequest,
) (*userinternalapi.GetUserResponse, error) {
	user, err := h.svc.FindUserByEmail(ctx, req.GetEmail())
	if err != nil {
		return nil, mapError(err)
	}
	return userResponse(user), nil
}

func userResponse(user domain.User) *userinternalapi.GetUserResponse {
	return &userinternalapi.GetUserResponse{
		UserId: user.ID.String(),
		Fields: &userinternalapi.UserFields{DisplayName: user.DisplayName, FirstName: user.FirstName, LastName: user.LastName, Email: user.Email},
	}
}

func (h *handler) UpdateUser(
	ctx context.Context,
	req *userinternalapi.UpdateUserRequest,
) (*userinternalapi.UpdateUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil || req.GetFields() == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if err = h.svc.UpdateUser(ctx, id, fieldsSpec(req.GetFields())); err != nil {
		return nil, mapError(err)
	}
	return &userinternalapi.UpdateUserResponse{}, nil
}

func (h *handler) DeleteUser(ctx context.Context, req *userinternalapi.DeleteUserRequest) (*userinternalapi.DeleteUserResponse, error) {
	id, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	if err = h.svc.DeleteUser(ctx, id); err != nil {
		return nil, mapError(err)
	}
	return &userinternalapi.DeleteUserResponse{}, nil
}

func fieldsSpec(fields *userinternalapi.UserFields) app.UserFieldsSpec {
	return app.UserFieldsSpec{
		DisplayName: fields.GetDisplayName(),
		FirstName:   fields.GetFirstName(),
		LastName:    fields.GetLastName(),
		Email:       fields.GetEmail(),
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidUser):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	default:
		return status.Error(codes.Internal, "user operation failed")
	}
}
