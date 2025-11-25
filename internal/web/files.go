package web

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/templates"
)

func UploadView(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, ok := GetSession(ctx)

		templates.Layout(templates.PageData{
			Authenticated: ok,
			Username:      s.Username,
			ProfilePath:   "",
			PageTitle:     "File upload",
			Place:         templates.PlaceUpload,
			Child:         templates.Upload(),
			Path:          r.URL,
			Err:           nil,
		}).Render(ctx, w)
	}
}

func Upload(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		s, _ := GetSession(ctx)
		err := r.ParseMultipartForm(MaxMemory)
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to read multipart form from request")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		description := r.Form.Get("description")
		file, header, err := r.FormFile("file")
		if err != nil {
			// TODO: better handle error.
			http.Error(w, "failed to get file", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read file body", http.StatusBadRequest)
			return
		}
		mime := http.DetectContentType(body)
		u, _, err := h.service.CreateFile(r.Context(), body, domain.FileMetadata{
			Name:       description,
			Filename:   header.Filename,
			Type:       "Image", // Handle it better.
			MimeType:   mime,
			SizeBytes:  int64(len(body)),
			UploaderId: s.UserID,
			Local:      true,
		})

		if err != nil {
			http.Error(w, "upload failed", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, u.String(), http.StatusSeeOther)
	}
}

func GetFile(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		digest := chi.URLParam(r, "digest")
		file, meta, err := h.service.GetFile(r.Context(), digest)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", meta.MimeType)
		w.Write(file)
	}
}
