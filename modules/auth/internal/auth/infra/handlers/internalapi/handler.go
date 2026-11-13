package internalapi

import (
	"context"
	"errors"
	"uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authinternalapi "auth/api/server/internalapi"
	"auth/internal/auth/app"
	"auth/internal/auth/domain"
)

type Handler struct {
	registration app.RegistrationParticipant
	authinternalapi.UnimplementedAuthServiceServer
}

func New(registration app.RegistrationParticipant) *Handler {
	return &Handler{registration: registration}
}

func (h *Handler) PrepareCredentialRegistration(ctx context.Context, request *authinternalapi.PrepareCredentialRegistrationRequest) (*authinternalapi.PrepareCredentialRegistrationResponse, error) {
	transactionID, transactionErr := uuid.Parse(request.GetTransactionId())
	userID, userErr := uuid.Parse(request.GetUserId())
	if transactionErr != nil || userErr != nil || request.GetLogin() == "" || request.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid registration transaction")
	}
	credentialID, err := h.registration.Prepare(ctx, transactionID, userID, request.GetLogin(), request.GetPassword())
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCredentialExists):
			return nil, status.Error(codes.AlreadyExists, "login already exists")
		case errors.Is(err, domain.ErrInvalidCredential), errors.Is(err, domain.ErrInvalidPassword):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "prepare credential registration")
		}
	}
	return &authinternalapi.PrepareCredentialRegistrationResponse{CredentialId: credentialID.String()}, nil
}

func (h *Handler) CommitCredentialRegistration(ctx context.Context, request *authinternalapi.RegistrationTransactionRequest) (*authinternalapi.RegistrationTransactionResponse, error) {
	transactionID, err := uuid.Parse(request.GetTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid registration transaction")
	}
	if err = h.registration.Commit(ctx, transactionID); err != nil {
		return nil, status.Error(codes.Unavailable, "commit credential registration")
	}
	return &authinternalapi.RegistrationTransactionResponse{}, nil
}

func (h *Handler) RollbackCredentialRegistration(ctx context.Context, request *authinternalapi.RegistrationTransactionRequest) (*authinternalapi.RegistrationTransactionResponse, error) {
	transactionID, err := uuid.Parse(request.GetTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid registration transaction")
	}
	if err = h.registration.Rollback(ctx, transactionID); err != nil {
		return nil, status.Error(codes.Unavailable, "rollback credential registration")
	}
	return &authinternalapi.RegistrationTransactionResponse{}, nil
}
