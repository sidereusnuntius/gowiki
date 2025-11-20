package web

import (
	"context"
	"errors"
	"net/http"

	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/service"
	"github.com/sidereusnuntius/wiki/templates"
)

func SignUp(s *Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)

			renderSignup(ctx, w, r, s.Config.ApprovalRequired, errors.New("failed to parse form body"))
			return
		}

		username := r.Form.Get("username")
		email := r.Form.Get("email")
		password := r.Form.Get("password")
		reason := r.Form.Get("reason")

		// Optional invitation code.
		invitation := r.PathValue("invitation")

		err = s.service.CreateUser(ctx, username, password, email, reason, false, invitation)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			renderSignup(ctx, w, r, s.Config.ApprovalRequired, err)
		}
	})
}

func renderSignup(ctx context.Context, w http.ResponseWriter, r *http.Request, approvalRequired bool, err error) {
	templates.Layout(templates.PageData{
		PageTitle: "Signup",
		Place:     templates.PlaceSignup,
		IsArticle: false,
		Child:     templates.SignUp("/signup", approvalRequired),
	}).Render(ctx, w)
}

func GetSignup(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderSignup(r.Context(), w, r, handler.Config.ApprovalRequired, nil)
	}
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
