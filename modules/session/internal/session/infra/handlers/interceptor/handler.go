package interceptor

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"uuid"

	"github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"
	"github.com/pkg/errors"

	"session/internal/session/app"
)

type Handler struct {
	interceptorgrpc.UnimplementedInterceptorServer
	sessions   app.SessionService
	cookieName string
}

func NewSessionInterceptor(sessions app.SessionService, cookieName string) *Handler {
	return &Handler{
		sessions:   sessions,
		cookieName: cookieName,
	}
}

func (h *Handler) Intercept(
	ctx context.Context,
	req *interceptorgrpc.InterceptRequest,
) (*interceptorgrpc.InterceptResponse, error) {
	if req.GetRequestHeaders() == nil {
		return continueResponse(), nil
	}
	cookieValue := ""
	for _, header := range req.GetRequestHeaders().GetHeaders() {
		if strings.EqualFold(header.GetKey(), "cookie") {
			request := http.Request{Header: http.Header{"Cookie": []string{header.GetValue()}}}
			if cookie, err := request.Cookie(h.cookieName); err == nil {
				cookieValue = cookie.Value
				break
			}
		}
	}
	sessionID, err := uuid.Parse(cookieValue)
	if err != nil {
		return unauthorized()
	}
	session, err := h.sessions.Validate(ctx, sessionID)
	if err != nil {
		return unauthorized()
	}

	return &interceptorgrpc.InterceptResponse{
		Result: &interceptorgrpc.InterceptResponse_ContinueResponse{
			ContinueResponse: &interceptorgrpc.ContinueResponse{
				HeaderMutation: &interceptorgrpc.HeaderMutation{
					Set: []*interceptorgrpc.HeaderValues{
						{Key: "X-Auth-User", Values: []string{session.UserID.String()}},
					},
				},
			},
		},
	}, nil
}

func continueResponse() *interceptorgrpc.InterceptResponse {
	return &interceptorgrpc.InterceptResponse{
		Result: &interceptorgrpc.InterceptResponse_ContinueResponse{
			ContinueResponse: &interceptorgrpc.ContinueResponse{},
		}}
}

func unauthorized() (*interceptorgrpc.InterceptResponse, error) {
	body, err := json.Marshal(map[string]string{"message": "authentication required"})
	return &interceptorgrpc.InterceptResponse{
		Result: &interceptorgrpc.InterceptResponse_ImmediateResponse{
			ImmediateResponse: &interceptorgrpc.ImmediateResponse{
				StatusCode: 401,
				Headers:    []*interceptorgrpc.Header{{Key: "Content-Type", Value: "application/json"}},
				Body:       body},
		},
	}, errors.WithStack(err)
}
