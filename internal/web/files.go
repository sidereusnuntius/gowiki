package web

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

func Upload(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(MaxMemory)
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to read multipart form from request")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		_, _, err = r.FormFile("file")
		
	}
}