package public

import (
	"context"
	"errors"
	"net/http"
	"time"
	"uuid"

	googleuuid "github.com/google/uuid"

	"auth/api/server/publicapi"
	"auth/internal/auth/app"
	"auth/internal/auth/config"
	"auth/internal/auth/domain"
)

type sessionContextKey struct{}

type handler struct {
	svc    app.AuthService
	cookie config.Cookie
}

func NewHandler(svc app.AuthService, cookie config.Cookie) publicapi.StrictServerInterface {
	return &handler{svc: svc, cookie: cookie}
}

func (h *handler) Login(ctx context.Context, req publicapi.LoginRequestObject) (publicapi.LoginResponseObject, error) {
	result, err := h.svc.Login(ctx, req.Body.Login, req.Body.Password, req.Body.RememberMe)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			return publicapi.Login401JSONResponse{Message: "invalid credentials"}, nil
		}
		if errors.Is(err, domain.ErrInvalidCredential) {
			return publicapi.Login400JSONResponse{Message: err.Error()}, nil
		}
		return publicapi.Login503JSONResponse{Message: "authentication unavailable"}, nil
	}
	return publicapi.Login200JSONResponse{
		Body:    authResult(result),
		Headers: publicapi.Login200ResponseHeaders{SetCookie: cookieString(h.sessionCookie(result))},
	}, nil
}

func (h *handler) Logout(ctx context.Context, _ publicapi.LogoutRequestObject) (publicapi.LogoutResponseObject, error) {
	err := h.svc.Logout(ctx, sessionID(ctx))
	if err != nil {
		return publicapi.Logout503JSONResponse{Message: "session unavailable"}, nil
	}
	return publicapi.Logout204Response{
		Headers: publicapi.Logout204ResponseHeaders{SetCookie: cookieString(h.expiredCookie())},
	}, nil
}

func (h *handler) ChangePassword(
	ctx context.Context,
	req publicapi.ChangePasswordRequestObject,
) (publicapi.ChangePasswordResponseObject, error) {
	id := sessionID(ctx)
	if id == uuid.Nil() {
		return unauthorizedPasswordResponse(), nil
	}
	err := h.svc.ChangePassword(ctx, id, req.Body.CurrentPassword, req.Body.NewPassword)
	switch {
	case err == nil:
		return publicapi.ChangePassword204Response{
			Headers: publicapi.ChangePassword204ResponseHeaders{
				SetCookie: cookieString(h.expiredCookie()),
			},
		}, nil
	case errors.Is(err, app.ErrSessionNotFound):
		return unauthorizedPasswordResponse(), nil
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrInvalidPassword):
		return publicapi.ChangePassword400JSONResponse{Message: err.Error()}, nil
	default:
		return publicapi.ChangePassword503JSONResponse{Message: "password change unavailable"}, nil
	}
}

func (h *handler) sessionCookie(result app.AuthResult) http.Cookie {
	//nolint:gosec // Secure is enabled by default and is configurable only for local HTTP tests.
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

func (h *handler) expiredCookie() http.Cookie {
	//nolint:gosec // Secure is enabled by default and is configurable only for local HTTP tests.
	return http.Cookie{
		Name:     h.cookie.Name,
		Domain:   h.cookie.Domain,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookie.Secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Expires:  time.Unix(1, 0).UTC(),
	}
}

func authResult(result app.AuthResult) publicapi.AuthResult {
	response := publicapi.AuthResult{
		UserID:        googleuuid.UUID(result.UserID),
		Authenticated: result.Authenticated,
	}
	if result.Authenticated {
		expiresAt := result.ExpiresAt
		response.ExpiresAt = &expiresAt
	}
	return response
}

func unauthorizedPasswordResponse() publicapi.ChangePasswordResponseObject {
	return publicapi.ChangePassword401JSONResponse{Message: "authentication required"}
}

func sessionID(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(sessionContextKey{}).(uuid.UUID)
	return id
}

func cookieString(cookie http.Cookie) string {
	return cookie.String()
}
