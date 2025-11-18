package web

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sidereusnuntius/wiki/templates"
)

func Profile(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, ok := GetSession(ctx)
		username := chi.URLParam(r, "username")
		domain := chi.URLParam(r, "domain")

		p, err := h.service.GetUserProfile(ctx, username, domain)
		if err != nil {
			//TODO
			log.Print(err)
			http.Error(w, "failed: handle error later", http.StatusInternalServerError)
			return
		}

		hrefs := map[templates.Place]string{
			templates.Read: r.URL.String(),
		}

		if ok {
			hrefs[templates.Edit] = r.URL.JoinPath("edit").String()
		}
		templates.Layout(templates.PageData{
			Authenticated: ok,
			Username:      u.Username,
			ProfilePath:   "TODO",
			PageTitle:     username,
			Place:         templates.PlaceProfile,
			Hrefs:         hrefs,
			IsArticle:     false,
			Child:         templates.Profile(p),
			FixedArticles: nil, // TODO
			Path:          r.URL,
			Err:           nil,
		}).Render(ctx, w)
	}
}
