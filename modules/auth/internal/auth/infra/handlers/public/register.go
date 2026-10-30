package public

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"uuid"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/go-chi/chi/v5"

	"auth/api/server/publicapi"
	"auth/internal/auth/app"
	"auth/internal/auth/config"
)

func Register(svc app.AuthService, logger *slog.Logger, cookie config.Cookie) (runtimehttp.RegisterFunc, error) {
	spec, err := publicapi.GetSwagger()
	if err != nil {
		return nil, err
	}
	validatorRouter, err := legacy.NewRouter(spec)
	if err != nil {
		return nil, err
	}
	return func(router chi.Router) error {
		router.Group(func(router chi.Router) {
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Cache-Control", "no-store")
					route, params, validationErr := validatorRouter.FindRoute(r)
					if validationErr == nil {
						validationErr = openapi3filter.ValidateRequest(
							r.Context(),
							&openapi3filter.RequestValidationInput{
								Request: r, PathParams: params, Route: route,
							},
						)
					}
					if validationErr != nil {
						writeValidationError(w, validationErr)
						return
					}
					ctx := r.Context()
					if value, cookieErr := r.Cookie(cookie.Name); cookieErr == nil {
						if id, parseErr := uuid.Parse(value.Value); parseErr == nil {
							ctx = context.WithValue(ctx, sessionContextKey{}, id)
						}
					}
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			})
			strict := publicapi.NewStrictHandlerWithOptions(
				NewHandler(svc, cookie),
				nil,
				publicapi.StrictHTTPServerOptions{
					ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
						logger.ErrorContext(r.Context(), "auth request failed", logging.Error(err))
						http.Error(w, "internal server error", http.StatusInternalServerError)
					},
				},
			)
			publicapi.HandlerFromMux(strict, router)
		})
		return nil
	}, nil
}

type validationError struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeValidationError(w http.ResponseWriter, validationErr error) {
	response := validationError{Message: "invalid request"}
	var schemaErr *openapi3.SchemaError
	if errors.As(validationErr, &schemaErr) {
		pointer := schemaErr.JSONPointer()
		if len(pointer) > 0 {
			response.Message = "validation failed"
			response.Fields = map[string]string{pointer[len(pointer)-1]: schemaErr.Reason}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(response)
}
