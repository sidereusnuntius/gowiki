package web

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/sidereusnuntius/wiki/web/templates"
)

const SessionKey = "user"

type Session struct {
	UserID int64
	AccountID int64
	Username string
}

type key struct{}

func GetSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(key{}).(Session)
	return s, ok
}

func SessionMiddleware(state State) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			zero := Session{}
			session := state.SessionManager.Load(r)
			var s Session
			err := session.GetObject(SessionKey, &s)
			if s != zero && err == nil {
				ctx := r.Context()
				ctx = context.WithValue(ctx, key{}, s)
				r = r.WithContext(ctx)
			}
	
			handler.ServeHTTP(w, r)
		})
	}
}

// TODO: idea: create a struct to represent a template, or just a function to facilitate the rendering of the standard, common layout.
func Login(state State) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		session := state.SessionManager.Load(r)
		err := r.ParseForm()
		if err != nil {
			//TODO: treat error.
			w.WriteHeader(http.StatusBadRequest)
			Render(
				r.Context(),
				w,
				templates.Login(LoginRoute),
				err,
			)
			return
		}
		
		user := r.Form.Get("user")
		password := r.Form.Get("password")
		u, authenticated, err := state.service.AuthenticateUser(ctx, user, password)
		if err != nil {
			//TODO: treat error.
			log.Print(err)
			w.WriteHeader(http.StatusBadRequest)
			Render(
				r.Context(),
				w,
				templates.Login(LoginRoute),
				err,
			)
			return
		}

		if !authenticated {
			w.WriteHeader(http.StatusBadRequest)
			Render(
				r.Context(),
				w,
				templates.Login(LoginRoute),
				errors.New("wrong credentials"),
			)
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
			Render(
				r.Context(),
				w,
				templates.Login(LoginRoute),
				errors.New("failed to create and load session"),
			)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})
}