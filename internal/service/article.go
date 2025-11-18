package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/db/queries"
	"github.com/sidereusnuntius/wiki/internal/validate"
)

// AlterArticle modifies the article with the given title or creates it if the article does not exist; the operation
// if the user does not have enough permissions to edit the wiki. If the operation succeeds, it returns the article's
// URL and a nil error.
func (s *Service) AlterArticle(ctx context.Context, title, summary, content string, userId int64) (string, error) {
	//TODO: deal with variations in the capitalization of the article title.
	//TODO: check if user has permission to edit the wiki and the article in question.
	articleId, ap, prev, err := s.DB.GetLastRevisionID(ctx, title)
	if err == nil {
		return ap, s.DB.UpdateArticle(ctx, prev, articleId, userId, summary, content)
	}

	if errors.Is(err, db.ErrNotFound) {
		ap, err = s.CreateArticle(ctx, title, summary, content, userId)
	}
	return ap, err

}

func (s *Service) GetLocalArticle(ctx context.Context, title string) (article queries.GetLocalArticleByTitleRow, err error) {
	title = RemoveDuplicateSpaces(title)
	err = validate.Title(title)
	if err != nil {
		return
	}

	article, err = s.DB.GetLocalArticle(ctx, title)
	return
}

func (s *Service) CreateArticle(ctx context.Context, title, summary, content string, userId int64) (string, error) {
	// TODO: validate title.
	title = RemoveDuplicateSpaces(title)
	summary = RemoveDuplicateSpaces(summary)

	err := validate.Title(title)
	if err != nil {
		return "", err
	}

	apId := s.Config.Url.JoinPath("a", title).String()
	article := queries.CreateArticleParams{
		Title:   title,
		Content: content,
		ApID:    apId,
		Url: sql.NullString{
			String: apId,
			Valid:  true,
		},
		Language:  s.Config.Language,
		MediaType: s.Config.MediaType,
	}

	diffs := s.FindDiff("", content)
	revision := queries.EditArticleParams{
		UserID: userId,
		Summary: sql.NullString{
			String: summary,
			Valid:  summary != "",
		},
		Diff:      diffs,
		Reviewed:  false,
		Published: true,
	}
	return article.ApID, s.DB.CreateLocalArticle(ctx, article, revision)
}

func (s *Service) FindDiff(text1, text2 string) string {
	diffs := s.DMP.DiffMain(text1, text2, false)
	return s.DMP.PatchToText(s.DMP.PatchMake(diffs))
}

// TODO: optimize.
func RemoveDuplicateSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
	// 	original := []byte(s)
	// 	l := len(original)
	// 	removed := make([]byte, 0, l)
	// 	var i int
	// 	var space bool
	// 	for original[i] == ' ' || original[i] == '\t' {
	// 		i++
	// 	}
	// 	for j := 0; i < l; i++ {
	// 		if space && (original[i] == ' ' && original[i] == '\t') {
	// 			continue
	// 		}
	// 		removed[j] = original[i]
	// 		j++
	// 		space = removed[i] == ' '
	// 	}

	// return string(removed)
}
