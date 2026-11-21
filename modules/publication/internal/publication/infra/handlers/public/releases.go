package public

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"publication/internal/publication/app/service"
	"strings"
	"uuid"

	"github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"

	"publication/internal/publication/app/model"
)

type releaseHandler struct{ releases *service.Releases }

func registerPageRoutes(router chi.Router, releases *service.Releases) {
	h := &releaseHandler{releases}
	router.Get("/p/{id}", h.page)
	router.Get("/p/{id}/assets/{file}", h.attachment)
}

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html><html lang="ru"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.Title}}</title><style>body{font:17px/1.6 system-ui;max-width:760px;margin:5vh auto;padding:0 20px;color:#24314b}article{overflow-wrap:anywhere}a{color:#3369c7}</style><main><h1>{{.Title}}</h1><article>{{.HTML}}</article>{{if .Assets}}<h2>Вложения</h2><ul>{{range .Assets}}<li><a href="/p/{{$.ID}}/assets/{{.FileID}}">{{.FileName}}</a></li>{{end}}</ul>{{end}}</main></html>`))

func (h *releaseHandler) page(w http.ResponseWriter, r *http.Request) {
	id, parseErr := uuid.Parse(chi.URLParam(r, "id"))
	if parseErr != nil {
		http.Error(w, "publication not found", http.StatusNotFound)
		return
	}
	release, attachments, err := h.releases.Public(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "publication not found", http.StatusNotFound)
		} else {
			http.Error(w, "publication unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	var raw bytes.Buffer
	if err = goldmark.Convert([]byte(release.Content), &raw); err != nil {
		http.Error(w, "render failed", http.StatusInternalServerError)
		return
	}
	clean := bluemonday.UGCPolicy().SanitizeBytes(raw.Bytes())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src 'self'; base-uri 'none'; form-action 'none'")
	_ = pageTemplate.Execute(w, struct {
		ID, Title string
		HTML      template.HTML
		Assets    []model.Attachment
	}{id.String(), release.Title, sanitizedHTML(clean), attachments})
}

func sanitizedHTML(clean []byte) template.HTML {
	// #nosec G203 -- clean is produced by bluemonday immediately before this call.
	return template.HTML(clean)
}

func (h *releaseHandler) attachment(w http.ResponseWriter, r *http.Request) {
	id, idErr := uuid.Parse(chi.URLParam(r, "id"))
	file, fileErr := uuid.Parse(chi.URLParam(r, "file"))
	if idErr != nil || fileErr != nil {
		http.Error(w, "attachment not found", http.StatusNotFound)
		return
	}
	meta, data, err := h.releases.DownloadAttachment(r.Context(), id, file)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "attachment not found", http.StatusNotFound)
		} else {
			http.Error(w, "attachment unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	contentType := meta.ContentType
	if !strings.HasPrefix(contentType, "image/png") && !strings.HasPrefix(contentType, "image/jpeg") && !strings.HasPrefix(contentType, "image/gif") && !strings.HasPrefix(contentType, "image/webp") {
		contentType = "application/octet-stream"
		w.Header().Set("Content-Disposition", "attachment")
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	// #nosec G705 -- attachment bytes are served with nosniff and non-image content is forced to attachment.
	_, _ = w.Write(data)
}
