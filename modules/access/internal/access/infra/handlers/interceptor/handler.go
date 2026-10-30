package interceptor

import (
	"context"
	"encoding/json"
	"strings"
	"uuid"

	interceptorgrpc "github.com/distributed-programming-2026/gateway/api/client/interceptorgrpc"

	"access/internal/access/app"
)

type Handler struct {
	interceptorgrpc.UnimplementedInterceptorServer
	svc app.ModuleAccessService
}

func New(svc app.ModuleAccessService) *Handler { return &Handler{svc: svc} }
func (h *Handler) Intercept(ctx context.Context, r *interceptorgrpc.InterceptRequest) (*interceptorgrpc.InterceptResponse, error) {
	if r.GetRequestHeaders() == nil {
		return allow(), nil
	}
	var user string
	for _, v := range r.GetRequestHeaders().GetHeaders() {
		if strings.EqualFold(v.GetKey(), "X-Auth-User") {
			user = v.GetValue()
			break
		}
	}
	userID, e := uuid.Parse(user)
	if e != nil {
		return deny(), nil
	}
	moduleID, e := uuid.Parse(r.GetContext().GetParameters()["moduleID"])
	if e != nil {
		return deny(), nil
	}
	ok, e := h.svc.CheckModuleAccess(ctx, userID, moduleID)
	if e != nil {
		return nil, e
	}
	if !ok {
		return deny(), nil
	}
	return allow(), nil
}
func allow() *interceptorgrpc.InterceptResponse {
	return &interceptorgrpc.InterceptResponse{Result: &interceptorgrpc.InterceptResponse_ContinueResponse{ContinueResponse: &interceptorgrpc.ContinueResponse{}}}
}
func deny() *interceptorgrpc.InterceptResponse {
	body, _ := json.Marshal(map[string]string{"message": "module access denied"})
	return &interceptorgrpc.InterceptResponse{Result: &interceptorgrpc.InterceptResponse_ImmediateResponse{ImmediateResponse: &interceptorgrpc.ImmediateResponse{StatusCode: 403, Headers: []*interceptorgrpc.Header{{Key: "Content-Type", Value: "application/json"}}, Body: body}}}
}
