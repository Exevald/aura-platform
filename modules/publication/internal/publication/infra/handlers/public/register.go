package public

import (
	"errors"
	"net/http"
	publicapi "publication/api/server/publicapi"
	"publication/internal/publication/app/service"

	"github.com/go-chi/chi/v5"

	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"
)

func Register(releases *service.Releases) (runtimehttp.RegisterFunc, error) {
	return func(router chi.Router) error {
		registerPageRoutes(router, releases)
		publicapi.HandlerWithOptions(publicapi.NewStrictHandler(&releaseAPI{releases}, nil), publicapi.ChiServerOptions{
			BaseRouter: router,
			ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				var missing *publicapi.RequiredHeaderError
				var invalid *publicapi.InvalidParamFormatError
				if errors.As(err, &missing) && isActorHeader(missing.ParamName) || errors.As(err, &invalid) && isActorHeader(invalid.ParamName) {
					http.Error(w, "authentication required", http.StatusUnauthorized)
					return
				}
				http.Error(w, "invalid request", http.StatusBadRequest)
			},
		})
		return nil
	}, nil
}

func isActorHeader(name string) bool {
	return name == "X-Auth-User" || name == "X-Workspace-ID"
}
