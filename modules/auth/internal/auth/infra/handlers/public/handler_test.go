package public

import (
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"

	"auth/internal/auth/app"
	"auth/internal/auth/config"
)

func TestSessionCookieUsesConfiguredDomain(t *testing.T) {
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	h := handler{cookie: config.Cookie{Name: "aura_session", Domain: "example.test", Secure: true}}

	cookie := h.sessionCookie(app.AuthResult{
		SessionID:  uuid.New(),
		ExpiresAt:  expiresAt,
		Persistent: true,
	})

	require.Equal(t, "example.test", cookie.Domain)
	require.Equal(t, "/", cookie.Path)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	require.Positive(t, cookie.MaxAge)
	require.Equal(t, expiresAt, cookie.Expires)
}

func TestExpiredCookieUsesSameScope(t *testing.T) {
	h := handler{cookie: config.Cookie{Name: "aura_session", Domain: "example.test", Secure: true}}

	cookie := h.expiredCookie()

	require.Equal(t, "example.test", cookie.Domain)
	require.Equal(t, "/", cookie.Path)
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Negative(t, cookie.MaxAge)
}
