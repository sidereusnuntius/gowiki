package web

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/sidereusnuntius/wiki/templates"
)

const SessionKey = "user"

type Session struct {
	UserID    int64
	AccountID int64
	Username  string
}

type key struct{}

func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(key{}).(Session)
	return s, ok
}

func AuthenticatedMiddleware(handler *Handler) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetSession(r.Context())
			if ok {
				handler.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Unauthenticated"))
			// Render a template.
		})
	}
}

func UnauthenticatedMiddleware(handler Handler) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := GetSession(r.Context())
			if ok {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			handler.ServeHTTP(w, r)
		})
	}
}

func Logout(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prev := r.URL.Query().Get("prev")
		s := handler.SessionManager.Load(r)
		err := s.Destroy(w)

		if err != nil {
			// TODO: do something! the question is: what?
		}

		if prev == "" {
			prev = "/"
		}
		http.Redirect(w, r, prev, http.StatusSeeOther)
	}
}

func SessionMiddleware(handler *Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			zero := Session{}
			session := handler.SessionManager.Load(r)
			var s Session
			err := session.GetObject(SessionKey, &s)
			if s != zero && err == nil {
				ctx := r.Context()
				ctx = context.WithValue(ctx, key{}, s)
				r = r.WithContext(ctx)
			}

			h.ServeHTTP(w, r)
		})
	}
}

// TODO: somehow get the page the user was on before clicking or being sent to the login page, so they can be redirected when they are logged in.
func Login(handler *Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := handler.SessionManager.Load(r)
		err := r.ParseForm()
		if err != nil {
			//TODO: treat error.
			w.WriteHeader(http.StatusBadRequest)
			renderLogin(
				r.Context(),
				w,
				r,
				err,
			)
			return
		}

		user := r.Form.Get("user")
		password := r.Form.Get("password")
		u, authenticated, err := handler.service.AuthenticateUser(ctx, user, password)
		if err != nil {
			//TODO: treat error.
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			renderLogin(
				r.Context(),
				w,
				r,
				err,
			)
			return
		}

		if !authenticated {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		err = session.PutObject(w, SessionKey, Session{
			u.UserID,
			u.AccountID,
			u.Username,
		})
		if err != nil {
			// Log error
			w.WriteHeader(http.StatusInternalServerError)
			renderLogin(
				r.Context(),
				w,
				r,
				errors.New("failed to create and load session"),
			)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})
}

// TODO: somehow get the page the user was on before clicking or being sent to the login page, so they can be redirected when they are logged in.
func GetLogin(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderLogin(r.Context(), w, r, nil)
	}
}

func renderLogin(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	_ = r.URL.Query().Get("prev")

	templates.Layout(templates.PageData{
		Authenticated: false,
		Username:      "",
		ProfilePath:   "",
		PageTitle:     "Login",
		Place:         templates.Auth,
		IsArticle:     false,
		Child:         templates.Login("/login"),
	}).Render(r.Context(), w)
}
