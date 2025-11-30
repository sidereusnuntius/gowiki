package wellknown

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/state"
)

type WebfingerLink struct {
	Rel string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

type WebfingerResponse struct {
	Subject string `json:"subject"`
	Links []WebfingerLink `json:"links"`
}

func Mount(state *state.State, r chi.Router) {
	r.Route("/.well-known/", func(r chi.Router) {
		r.Get("/webfinger", WebfingerEndpoint(state))
	})
}

func WebfingerEndpoint(state *state.State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resource := r.URL.Query().Get("resource")
		uri, err := url.Parse(strings.Replace(resource, "acct:", "acct://", 1))
		if err != nil {
			http.Error(w, "failed to parse resource", http.StatusBadRequest)
			return
		}

		apId, err := state.DB.GetUserApId(r.Context(), uri.User.Username())
		if err != nil {
			http.Error(w, "", handleErr(err))
			return
		}

		res := WebfingerResponse{
			Subject: resource,
			Links: []WebfingerLink{
				{Rel: "self", Type: "application/activity+json", Href: apId.String()},
			},
		}
		encoder := json.NewEncoder(w)
		
		if err = encoder.Encode(res); err != nil {
			log.Error().Err(err).Msg("unable to marshal webfinger response")
			http.Error(w, "", http.StatusInternalServerError)
		}
	}
}

func handleErr(err error) int {
	switch {
	case errors.Is(err, db.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}