package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/templates"
)

const MaxMemory = 64 * 1024

// ArticleHistory renders a template displaying all edits made to an article, if such article exists.
func ArticleHistory(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, ok := GetSession(ctx)
		title := chi.URLParam(r, "title")
		//page := r.PathValue.Get("after")
		list, err := handler.service.GetRevisionList(ctx, title)
		if err != nil {
			//TODO: handle error
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		history := r.URL.String()
		// TODO: store article URL in database, use it to generate paths.
		path, _ := url.Parse("/a/" + title)
		templates.Layout(templates.PageData{
			Authenticated: ok,
			Username:      u.Username,
			ProfilePath:   "TODO",
			PageTitle:     "Revision history",
			Place:         templates.History,
			Path:          r.URL,
			Hrefs: map[templates.Place]string{
				templates.Read:    path.String(),
				templates.Edit:    path.JoinPath("edit").String(),
				templates.History: history,
			},
			IsArticle: false,
			Child:     templates.Revisions(title, list),
		}).Render(ctx, w)
	}
}

// EditArticle renders the article editing screen, showing a textarea populated with the article's text and
// a summary of the edit.
func EditArticle(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: should I verify whether the user is logged in, or should I just assume that? I think I can't, since I need to use the user's username on the template.
		// TODO: verify whether article exists.
		// TODO: a revision can be based on a revision that is not the latest.
		ctx := r.Context()
		u, ok := GetSession(ctx) // Validate ok
		if !ok {

		}

		var newarticle bool
		title := chi.URLParam(r, "title")
		article, err := handler.service.GetLocalArticle(ctx, title)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				newarticle = true
			} else {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		err = r.ParseMultipartForm(MaxMemory)

		var content, summary string
		if err == nil {
			content = r.Form.Get("content")
			summary = r.Form.Get("summary")
		}

		var preview string
		if content != "" {
			preview = content
		}

		if content == "" {
			content = article.Content
		}

		//TODO: parse content to produce preview

		edit := r.URL.String()
		// TODO: store article URL in database, use it to generate paths.
		path, _ := url.Parse("/a/" + title)
		hrefs := map[templates.Place]string{
			templates.Edit: edit,
		}
		if !newarticle {
			hrefs[templates.Read] = path.String()
			hrefs[templates.History] = path.JoinPath("history").String()
		}

		templates.Layout(templates.PageData{
			Authenticated: ok,
			Username:      u.Username,
			ProfilePath:   "TODO",
			PageTitle:     "Editing " + article.Title,
			Place:         templates.Edit,
			Path:          r.URL,
			Hrefs:         hrefs,
			IsArticle:     false,
			Child:         templates.Editor(path.String(), edit, title, summary, preview, content),
		}).Render(ctx, w)
	}
}

func GetArticle(handler *Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, ok := GetSession(ctx)
		title := chi.URLParam(r, "title")
		article, err := handler.service.GetLocalArticle(ctx, title)

		// TODO: deal with the case in which the article has not been created, which should redirect to the editor.
		if err != nil {
			// TODO: render template
			if errors.Is(err, db.ErrNotFound) {
				templates.Layout(templates.PageData{
					Authenticated: ok,
					Username:      u.Username,
					ProfilePath:   "TODO",
					PageTitle:     title,
					Place:         templates.Read,
					Path:          r.URL,
					IsArticle:     false,
					Child:         templates.NonexistingArticle(r.URL.JoinPath("edit").String(), title),
				}).Render(ctx, w)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
			}
			return
		}

		// Sanitize content!
		templates.Layout(templates.PageData{
			Authenticated: ok,
			Username:      u.Username,
			ProfilePath:   "TODO",
			PageTitle:     article.Title,
			Place:         templates.Read,
			Path:          r.URL,
			Hrefs: map[templates.Place]string{
				templates.Read:    r.URL.String(),
				templates.Edit:    r.URL.JoinPath("edit").String(),
				templates.History: r.URL.JoinPath("history").String(),
			},
			IsArticle: true,
			Article: templates.ArticleData{
				Title:    article.Title,
				Domain:   "", //TODO
				URL:      r.URL,
				Content:  article.Content,
				Language: article.Language,
				License:  "", //TODO
			},
		}).Render(ctx, w)
	}
}

func PostArticle(handler *Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement a payload limit.
		ctx := r.Context()
		session, _ := GetSession(ctx)

		err := r.ParseMultipartForm(MaxMemory)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Print(w, "failed to parse form body")
			return
		}

		title := chi.URLParam(r, "title")

		summary := r.Form.Get("summary")
		content := r.Form.Get("content")
		//prev := r.Form.Get("")
		id, err := handler.service.AlterArticle(ctx, title, summary, content, session.UserID)
		if err == nil {
			http.Redirect(w, r, id.String(), http.StatusSeeOther)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	})
}

func Article(handler *Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
