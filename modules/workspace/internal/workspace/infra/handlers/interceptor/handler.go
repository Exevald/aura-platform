package interceptor

import (
	"context"
	"encoding/json"
	"strings"
	"uuid"
	"workspace/internal/workspace/app"

	interceptorgrpc "github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"
)

type Handler struct {
	interceptorgrpc.UnimplementedInterceptorServer
	svc *app.Service
}

func New(s *app.Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Intercept(ctx context.Context, r *interceptorgrpc.InterceptRequest) (*interceptorgrpc.InterceptResponse, error) {
	var user, workspace string
	if r.GetRequestHeaders() != nil {
		for _, v := range r.GetRequestHeaders().GetHeaders() {
			switch {
			case strings.EqualFold(v.GetKey(), "X-Auth-User"):
				user = v.GetValue()
			case strings.EqualFold(v.GetKey(), "X-Workspace-ID"):
				workspace = v.GetValue()
			}
		}
	}
	u, e := uuid.Parse(user)
	if e != nil {
		return deny(), nil
	}
	w, e := uuid.Parse(workspace)
	if e != nil {
		return deny(), nil
	}
	if _, _, e = h.svc.Check(ctx, u, w); e != nil {
		return deny(), nil
	}
	return allow(), nil
}

func allow() *interceptorgrpc.InterceptResponse {
	return &interceptorgrpc.InterceptResponse{Result: &interceptorgrpc.InterceptResponse_ContinueResponse{ContinueResponse: &interceptorgrpc.ContinueResponse{}}}
}

func deny() *interceptorgrpc.InterceptResponse {
	body, _ := json.Marshal(map[string]string{"message": "workspace access denied"})
	return &interceptorgrpc.InterceptResponse{Result: &interceptorgrpc.InterceptResponse_ImmediateResponse{ImmediateResponse: &interceptorgrpc.ImmediateResponse{StatusCode: 403, Headers: []*interceptorgrpc.Header{{Key: "Content-Type", Value: "application/json"}}, Body: body}}}
}
