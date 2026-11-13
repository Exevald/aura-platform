package public

import (
	"editor/internal/editor/domain/model"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"uuid"

	publicapi "editor/api/server/publicapi"
	"editor/internal/editor/app"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/go-chi/chi/v5"

	runtimehttp "github.com/distributed-programming-2026/lib/runtime/http"
)

func Register(svc app.EditorService, logger *slog.Logger) (runtimehttp.RegisterFunc, error) {
	spec, err := publicapi.GetSwagger()
	if err != nil {
		return nil, err
	}
	validatorRouter, err := legacy.NewRouter(spec)
	if err != nil {
		return nil, err
	}
	return func(router chi.Router) error {
		registerAttachmentRoutes(router, svc)
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

func registerAttachmentRoutes(router chi.Router, svc app.EditorService) {
	base := "/api/editor/v1/documents/{documentID}/attachments"
	router.Get(base, func(w http.ResponseWriter, r *http.Request) {
		actor, doc, workspace, ok := attachmentContext(w, r, true)
		if !ok {
			return
		}
		items, e := svc.ListAttachments(r.Context(), actor, workspace, doc)
		if !attachmentError(w, e) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	})
	router.Post(base, func(w http.ResponseWriter, r *http.Request) {
		actor, doc, workspace, ok := attachmentContext(w, r, true)
		if !ok {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
		file, header, e := r.FormFile("file")
		if e != nil {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, e := io.ReadAll(file)
		if e != nil {
			http.Error(w, "invalid file", http.StatusBadRequest)
			return
		}
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		value, e := svc.UploadAttachment(r.Context(), actor, workspace, doc, header.Filename, contentType, data)
		if !attachmentError(w, e) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(value)
	})
	router.Get(base+"/{fileID}/content", func(w http.ResponseWriter, r *http.Request) {
		actor, doc, workspace, ok := attachmentContext(w, r, true)
		if !ok {
			return
		}
		file, e := uuid.Parse(chi.URLParam(r, "fileID"))
		if e != nil {
			http.Error(w, "invalid file", http.StatusBadRequest)
			return
		}
		meta, data, e := svc.DownloadAttachment(r.Context(), actor, workspace, doc, file)
		if !attachmentError(w, e) {
			return
		}
		w.Header().Set("Content-Type", meta.ContentType)
		w.Header().Set("Content-Disposition", mimeDisposition(meta.FileName))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		// #nosec G705 -- attachments are downloaded with nosniff, not rendered inline.
		_, _ = w.Write(data)
	})
	router.Delete(base+"/{fileID}", func(w http.ResponseWriter, r *http.Request) {
		actor, doc, workspace, ok := attachmentContext(w, r, true)
		if !ok {
			return
		}
		file, e := uuid.Parse(chi.URLParam(r, "fileID"))
		if e != nil {
			http.Error(w, "invalid file", http.StatusBadRequest)
			return
		}
		if !attachmentError(w, svc.DeleteAttachment(r.Context(), actor, workspace, doc, file)) {
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func attachmentContext(w http.ResponseWriter, r *http.Request, workspaceRequired bool) (actorID, documentID, workspaceID uuid.UUID, ok bool) {
	actor, e := uuid.Parse(r.Header.Get("X-Auth-User"))
	if e != nil {
		http.Error(w, "invalid user", http.StatusBadRequest)
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), false
	}
	doc, e := uuid.Parse(chi.URLParam(r, "documentID"))
	if e != nil {
		http.Error(w, "invalid document", http.StatusBadRequest)
		return uuid.Nil(), uuid.Nil(), uuid.Nil(), false
	}
	workspace := uuid.Nil()
	if workspaceRequired {
		workspace, e = uuid.Parse(r.Header.Get("X-Workspace-ID"))
		if e != nil {
			http.Error(w, "invalid workspace", http.StatusBadRequest)
			return uuid.Nil(), uuid.Nil(), uuid.Nil(), false
		}
	}
	return actor, doc, workspace, true
}

func attachmentError(w http.ResponseWriter, e error) bool {
	if e == nil {
		return true
	}
	switch {
	case errors.Is(e, model.ErrPermissionDenied):
		http.Error(w, e.Error(), http.StatusForbidden)
	case errors.Is(e, model.ErrDocumentNotFound):
		http.Error(w, "not found", http.StatusNotFound)
	default:
		http.Error(w, "attachment service unavailable", http.StatusServiceUnavailable)
	}
	return false
}

func mimeDisposition(name string) string {
	name = strings.ReplaceAll(strings.ReplaceAll(name, "\"", ""), "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	return `attachment; filename="` + name + `"`
}
