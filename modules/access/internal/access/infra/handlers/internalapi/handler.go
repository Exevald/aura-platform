package internalapi

import (
	"context"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	accessinternalapi "access/api/server/internalapi"
	"access/internal/access/app"
)

type Handler struct {
	accessinternalapi.UnimplementedModuleAccessServiceServer
	svc          app.ModuleAccessService
	registration app.RegistrationParticipant
}

func New(svc app.ModuleAccessService, registration app.RegistrationParticipant) *Handler {
	return &Handler{svc: svc, registration: registration}
}

func (h *Handler) GrantModuleAccess(ctx context.Context, r *accessinternalapi.GrantModuleAccessRequest) (*accessinternalapi.GrantModuleAccessResponse, error) {
	u, m, e := ids(r.GetUserId(), r.GetModuleId())
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	if e = h.svc.GrantModuleAccess(ctx, u, m); e != nil {
		return nil, status.Error(codes.Internal, "grant failed")
	}
	return &accessinternalapi.GrantModuleAccessResponse{}, nil
}

func (h *Handler) PrepareModuleAccessRegistration(ctx context.Context, r *accessinternalapi.PrepareModuleAccessRegistrationRequest) (*accessinternalapi.RegistrationTransactionResponse, error) {
	transactionID, transactionErr := uuid.Parse(r.GetTransactionId())
	userID, moduleID, idsErr := ids(r.GetUserId(), r.GetModuleId())
	if transactionErr != nil || idsErr != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid registration transaction")
	}
	if err := h.registration.Prepare(ctx, transactionID, userID, moduleID); err != nil {
		return nil, status.Error(codes.Internal, "prepare registration failed")
	}
	return &accessinternalapi.RegistrationTransactionResponse{}, nil
}

func (h *Handler) CommitModuleAccessRegistration(ctx context.Context, r *accessinternalapi.RegistrationTransactionRequest) (*accessinternalapi.RegistrationTransactionResponse, error) {
	transactionID, err := uuid.Parse(r.GetTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid registration transaction")
	}
	if err = h.registration.Commit(ctx, transactionID); err != nil {
		return nil, status.Error(codes.Unavailable, "commit registration transaction")
	}
	return &accessinternalapi.RegistrationTransactionResponse{}, nil
}

func (h *Handler) RollbackModuleAccessRegistration(ctx context.Context, r *accessinternalapi.RegistrationTransactionRequest) (*accessinternalapi.RegistrationTransactionResponse, error) {
	transactionID, err := uuid.Parse(r.GetTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid registration transaction")
	}
	if err = h.registration.Rollback(ctx, transactionID); err != nil {
		return nil, status.Error(codes.Unavailable, "rollback registration transaction")
	}
	return &accessinternalapi.RegistrationTransactionResponse{}, nil
}

func (h *Handler) RevokeModuleAccess(ctx context.Context, r *accessinternalapi.RevokeModuleAccessRequest) (*accessinternalapi.RevokeModuleAccessResponse, error) {
	u, m, e := ids(r.GetUserId(), r.GetModuleId())
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	if e = h.svc.RevokeModuleAccess(ctx, u, m); e != nil {
		return nil, status.Error(codes.Internal, "revoke failed")
	}
	return &accessinternalapi.RevokeModuleAccessResponse{}, nil
}

func (h *Handler) CheckModuleAccess(ctx context.Context, r *accessinternalapi.CheckModuleAccessRequest) (*accessinternalapi.CheckModuleAccessResponse, error) {
	u, m, e := ids(r.GetUserId(), r.GetModuleId())
	if e != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ids")
	}
	allowed, e := h.svc.CheckModuleAccess(ctx, u, m)
	if e != nil {
		return nil, status.Error(codes.Internal, "check failed")
	}
	return &accessinternalapi.CheckModuleAccessResponse{Allowed: allowed}, nil
}

func ids(user, module string) (userID, moduleID uuid.UUID, err error) {
	u, e := uuid.Parse(user)
	if e != nil {
		return uuid.Nil(), uuid.Nil(), e
	}
	m, e := uuid.Parse(module)
	return u, m, e
}
