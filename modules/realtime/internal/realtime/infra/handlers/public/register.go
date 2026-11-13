package public

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"uuid"

	publicapi "realtime/api/server/publicapi"
	"realtime/internal/realtime/app"

	"github.com/distributed-programming-2026/go-sdk/pkg/logging"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

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
		router.Get("/api/realtime/v1/ws", websocketHandler(svc, logger))
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

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	return err == nil && parsed.Host == r.Host
}}

func websocketHandler(svc *app.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(r.Header.Get("X-Auth-User"))
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		workspaceID, err := uuid.Parse(r.Header.Get("X-Workspace-ID"))
		if err != nil {
			http.Error(w, "workspace required", http.StatusBadRequest)
			return
		}
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		subscription, cancel := svc.Subscribe(userID, workspaceID)
		defer cancel()
		defer connection.Close()
		connection.SetReadLimit(4096)
		_ = connection.SetReadDeadline(time.Now().Add(60 * time.Second))
		connection.SetPongHandler(func(string) error { return connection.SetReadDeadline(time.Now().Add(60 * time.Second)) })
		readDone := make(chan struct{})
		go func() {
			defer close(readDone)
			for {
				if _, _, e := connection.ReadMessage(); e != nil {
					return
				}
			}
		}()
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-readDone:
				return
			case <-ticker.C:
				if err = connection.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second)); err != nil {
					return
				}
			case message, ok := <-subscription.Messages:
				if !ok {
					return
				}
				payload := struct {
					Type    string          `json:"type"`
					Payload json.RawMessage `json:"payload"`
				}{message.Type, json.RawMessage(message.Payload)}
				if err = connection.WriteJSON(payload); err != nil {
					logger.DebugContext(r.Context(), "websocket write failed", logging.Error(err))
					return
				}
			}
		}
	}
}
