package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/alexedwards/scs"
	"github.com/go-chi/chi/v5"
	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/service"
	"github.com/sidereusnuntius/wiki/web/templates"
)

const (
	LoginRoute = "/login"
	SignUpRoute = "/signup"
)

type State struct {
	Config *config.Configuration
	service *service.Service
	SessionManager *scs.Manager
}

func Render(ctx context.Context, w http.ResponseWriter, template templ.Component, err error) {
	u, _ := GetSession(ctx)

	w.Header().Add("Content-Type", "text/html")
	templates.
		Layout(template, SignUpRoute, LoginRoute, u.Username, err).
		Render(ctx, w)
}

func RenderHandler(template templ.Component) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Render(r.Context(), w, template, nil)
	})
}

func Mux(config *config.Configuration, service *service.Service, manager *scs.Manager) http.Handler {
	state := State{
		Config: config,
		service: service,
		SessionManager: manager,
	}
	mux := chi.NewMux()
	mux.Use(SessionMiddleware(state))

	mux.Get(LoginRoute, RenderHandler(templates.Login(LoginRoute)))
	mux.Post(LoginRoute, Login(state))

	mux.Post(SignUpRoute, SignUp(&state))
	mux.Get(SignUpRoute, RenderHandler(templates.SignUp(SignUpRoute, config.ApprovalRequired)))
	return mux
}

func SignUp(s *State) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			Render(
				r.Context(),
				w,
				templates.SignUp(SignUpRoute, s.Config.ApprovalRequired),
				errors.New("failed to parse form body"),
			)
			return
		}

		username := r.Form.Get("username")
		email := r.Form.Get("email")
		password := r.Form.Get("password")
		reason := r.Form.Get("reason")
		
		// Optional invitation code.
		invitation := r.PathValue("invitation")

		err = s.service.CreateUser(r.Context(), username, password, email, reason, false, invitation)
		if err != nil {
			
			w.WriteHeader(http.StatusInternalServerError)
			Render(
				r.Context(),
				w,
				templates.SignUp(SignUpRoute, s.Config.ApprovalRequired),
				err,
			)
		}
	})
}

func GetCode(w http.ResponseWriter, err error) int {
	switch {
	case errors.Is(err, db.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, service.ErrInvalidInput):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}