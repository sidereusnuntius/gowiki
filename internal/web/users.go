package web

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/templates"
)

func (h *Handler) profile(ctx context.Context, w http.ResponseWriter, r *http.Request, name, host string) error {
	u, ok := GetSession(ctx)

	p, err := h.service.GetProfile(ctx, name, host)
	if err != nil {
		return err
	}

	hrefs := map[templates.Place]string{
		templates.Read: r.URL.String(),
	}

	isAdmin, err := h.service.IsAdmin(ctx, u.AccountID)
	if err != nil {
		return err
	}

	var editable bool
	if ok && p.Name == u.Username {
		editable = true
	}
	if ok && p.Name == h.Config.Name && isAdmin {
		editable = true
	}

	if editable {
		hrefs[templates.Edit] = r.URL.JoinPath("edit").String()
	}
	return templates.Layout(templates.PageData{
		Authenticated: ok,
		Username:      u.Username,
		ProfilePath:   r.URL.String(),
		PageTitle:     p.Name,
		Place:         templates.PlaceProfile,
		Hrefs:         hrefs,
		IsArticle:     false,
		Child:         templates.Profile(p),
		FixedArticles: nil, // TODO
		Path:          r.URL,
		Err:           nil,
	}).Render(ctx, w)
}

func Profile(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		host := chi.URLParam(r, "host")
		if err := h.profile(r.Context(), w, r, name, host); err != nil {
			log.Error().Err(err).Msg("error displaying profile")
		}
	}
}
