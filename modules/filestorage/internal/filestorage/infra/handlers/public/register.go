package public

import (
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"uuid"

	publicapi "filestorage/api/server/publicapi"
	"filestorage/internal/filestorage/app"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/go-chi/chi/v5"

	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"
)

func Register(svc *app.Service, logger *slog.Logger) (runtimehttp.RegisterFunc, error) {
	spec, err := publicapi.GetSwagger()
	if err != nil {
		return nil, err
	}
	validatorRouter, err := legacy.NewRouter(spec)
	if err != nil {
		return nil, err
	}
	return func(router chi.Router) error {
		router.Post("/api/files/v1/files", upload(svc))
		router.Get("/api/files/v1/files/{fileID}/content", download(svc))
		router.Group(func(router chi.Router) {
			router.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					route, params, err := validatorRouter.FindRoute(r)
					if err == nil {
						err = openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
							Request:    r,
							PathParams: params,
							Route:      route,
						})
					}
					if err != nil {
						http.Error(w, "invalid request", http.StatusBadRequest)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			strict := publicapi.NewStrictHandlerWithOptions(NewHandler(svc), nil, publicapi.StrictHTTPServerOptions{
				ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					logger.ErrorContext(r.Context(), "request failed", logging.Error(err))
					http.Error(w, "internal server error", http.StatusInternalServerError)
				},
			})
			publicapi.HandlerFromMux(strict, router)
		})
		return nil
	}, nil
}

func upload(svc *app.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, workspaceID, ok := requestContext(w, r)
		if !ok {
			return
		}
		const maxUploadBytes = 32 << 20
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
		// #nosec G120 -- MaxBytesReader bounds the entire multipart body above.
		if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
			http.Error(w, "invalid multipart upload", http.StatusBadRequest)
			return
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		value, err := svc.Upload(r.Context(), workspaceID, userID, header.Filename, contentType, header.Size, file)
		if err != nil {
			http.Error(w, "upload failed", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(toAPI(value))
	}
}

func download(svc *app.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, workspaceID, ok := requestContext(w, r)
		if !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "fileID"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		value, body, err := svc.Download(r.Context(), workspaceID, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer body.Close()
		w.Header().Set("Content-Type", value.ContentType)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": value.FileName}))
		w.Header().Set("X-Content-SHA256", value.ChecksumSHA256)
		_, _ = io.Copy(w, body)
	}
}

func requestContext(w http.ResponseWriter, r *http.Request) (userID, workspaceID uuid.UUID, ok bool) {
	userID, err := uuid.Parse(r.Header.Get("X-Auth-User"))
	if err != nil {
		http.Error(w, "invalid user context", http.StatusBadRequest)
		return uuid.Nil(), uuid.Nil(), false
	}
	workspaceID, err = uuid.Parse(r.Header.Get("X-Workspace-ID"))
	if err != nil {
		http.Error(w, "invalid workspace context", http.StatusBadRequest)
		return uuid.Nil(), uuid.Nil(), false
	}
	return userID, workspaceID, true
}
