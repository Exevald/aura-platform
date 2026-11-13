package public

import (
	"context"
	"errors"
	"net/http"
	"time"

	googleuuid "github.com/google/uuid"

	"registerer/api/server/publicapi"
	"registerer/internal/registerer/app"
	"registerer/internal/registerer/config"
)

type handler struct {
	svc    *app.Service
	cookie config.Cookie
}

func NewHandler(svc *app.Service, cookie config.Cookie) publicapi.StrictServerInterface {
	return &handler{svc: svc, cookie: cookie}
}

func (h *handler) Register(
	ctx context.Context,
	request publicapi.RegisterRequestObject,
) (publicapi.RegisterResponseObject, error) {
	result, err := h.svc.Register(ctx, app.RegisterSpec{
		Login:      request.Body.Login,
		Password:   request.Body.Password,
		RememberMe: request.Body.RememberMe,
		Profile: app.ProfileSpec{
			DisplayName: request.Body.DisplayName,
			FirstName:   request.Body.FirstName,
			LastName:    request.Body.LastName,
			Email:       string(request.Body.Email),
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, app.ErrLoginTaken), errors.Is(err, app.ErrEmailTaken):
			return publicapi.Register409JSONResponse{Message: err.Error()}, nil
		case errors.Is(err, app.ErrInvalidRegistration):
			return publicapi.Register400JSONResponse{Message: err.Error()}, nil
		default:
			return publicapi.Register503JSONResponse{Message: "registration unavailable"}, nil
		}
	}
	response := publicapi.Register201JSONResponse{
		Body: publicapi.RegisterResult{
			UserID:        googleuuid.UUID(result.UserID),
			Authenticated: result.Authenticated,
		},
	}
	if result.Authenticated {
		expiresAt := result.ExpiresAt
		response.Body.ExpiresAt = &expiresAt
		sessionCookie := h.sessionCookie(result)
		cookieValue := sessionCookie.String()
		response.Headers.SetCookie = &cookieValue
	}
	return response, nil
}

func (h *handler) sessionCookie(result app.RegisterResult) http.Cookie {
	//nolint:gosec
	cookie := http.Cookie{
		Name:     h.cookie.Name,
		Value:    result.SessionID.String(),
		Domain:   h.cookie.Domain,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	if result.Persistent {
		cookie.Expires = result.ExpiresAt
		cookie.MaxAge = max(1, int(time.Until(result.ExpiresAt).Seconds()))
	}
	return cookie
}
