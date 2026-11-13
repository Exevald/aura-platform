package public

import (
	"log/slog"
	"net/http"

	publicapi "access/api/server/publicapi"
	"access/internal/access/app"
	clientinfra "access/internal/access/infra/client"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/go-chi/chi/v5"

	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"
)

func Register(svc app.ModuleAccessService, groups clientinfra.UserGroup, logger *slog.Logger, registries ...clientinfra.AppRegistry) (runtimehttp.RegisterFunc, error) {
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
					route, params, validationErr := validatorRouter.FindRoute(r)
					if validationErr == nil {
						validationErr = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{Request: r, PathParams: params, Route: route})
					}
					if validationErr != nil {
						http.Error(w, "invalid request", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			strict := publicapi.NewStrictHandlerWithOptions(NewHandler(svc, groups, registries...), nil, publicapi.StrictHTTPServerOptions{ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				logger.ErrorContext(r.Context(), "access request failed", logging.Error(err))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}})
			publicapi.HandlerFromMux(strict, router)
		})
		return nil
	}, nil
}
