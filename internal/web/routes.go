package web

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Mount(r chi.Router) {
	authenticated := AuthenticatedMiddleware(h)
	r.Use(SessionMiddleware(h))

	r.Route("/", func(r chi.Router) {
		//r.Use(UnauthenticatedMiddleware(handler))
		r.Get(LoginRoute, GetLogin(h))
		r.Post(LoginRoute, Login(h))

		r.Post(SignUpRoute, SignUp(h))
		r.Get(SignUpRoute, GetSignup(h))
		r.Get("/logout", Logout(h))
	})
	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	chi.Chain()
	// 	ServeHTTP(w, r)
	// })

	r.Get("/@{name}", Profile(h))
	r.Get("/@{name}@{host}", Profile(h))

	r.Route("/a/{title}", func(r chi.Router) {
		r.Post("/", PostArticle(h))
		r.Get("/", GetArticle(h))
		r.Handle("/edit", authenticated(EditArticle(h)))
		r.Get("/history", ArticleHistory(h))
	})

	r.Route("/a/{title}@{author}@{host}", func(r chi.Router) {
		r.Post("/", PostArticle(h))
		r.Get("/", GetArticle(h))
		r.Handle("/edit", authenticated(EditArticle(h)))
		r.Get("/history", ArticleHistory(h))
	})

	r.Route("/f", func(r chi.Router) {
		r.Get("/upload", authenticated(UploadView(h)))
		r.Post("/upload", authenticated(Upload(h)))
		r.Get("/{digest}", GetFile(h))
	})

	h.MountStaticRoutes(r)
}

func (h *Handler) MountStaticRoutes(r chi.Router) {
	wd, _ := os.Getwd()
	wd = filepath.Join(wd, h.Config.StaticDir)
	f := os.DirFS(wd)

	fileServer := http.FileServer(http.FS(f))
	r.Handle("/static/{name}", http.StripPrefix(
		"/static/",
		fileServer,
	))
}
