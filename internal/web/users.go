package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/templates"
)

func (h *Handler) profile(ctx context.Context, w http.ResponseWriter, r *http.Request, name, host string) error {
	s, ok := GetSession(ctx)

	p, err := h.service.GetProfile(ctx, name, host)
	if err != nil {
		return err
	}

	hrefs := map[templates.Place]string{
		templates.Read: r.URL.String(),
	}

	var isAdmin, followed bool
	if ok {
		if isAdmin, err = h.service.IsAdmin(ctx, s.AccountID); err != nil {
			return fmt.Errorf("cannot verify admin status for user %s: %w", s.Username, err)
		}
	}

	if isAdmin {
		var iri *url.URL
		if iri, err = h.service.GetActorIRI(ctx, name, host); err != nil {
			return err
		}
		
		if followed, err = h.service.Follows(ctx, h.Config.Url, iri); err != nil {
			return err
		}
	}

	var editable bool
	if ok && p.Name == s.Username {
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
		Username:      s.Username,
		ProfilePath:   r.URL.String(),
		PageTitle:     p.Name,
		Place:         templates.PlaceProfile,
		Hrefs:         hrefs,
		IsArticle:     false,
		Child:         templates.Profile(p, r.URL, isAdmin, followed),
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

// InstanceFollow allows an administrator to follow another actor, typically another wiki, on behalf of the wiki's
// instance actor. This is the way two wikis "peer", making them send updates to each other.
func InstanceFollow(h *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := chi.URLParam(r, "name")
		host := chi.URLParam(r, "host")

		last := r.URL.Query().Get("last")
		

		iri, err := h.service.GetActorIRI(ctx, name, host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = h.service.FollowRemote(r.Context(), h.Config.Url, iri)
		if err != nil {
			http.Error(w, "failed to follow actor", http.StatusInternalServerError)
			return
		}

		if last == "" {
			return
		}
		http.Redirect(w, r, last, http.StatusSeeOther)
	}
}