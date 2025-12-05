package core

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/validate"
)

// AlterArticle modifies the article with the given title or creates it if the article does not exist; the operation
// if the user does not have enough permissions to edit the wiki. If the operation succeeds, it returns the article's
// URL and a nil error.
func (s *AppService) AlterArticle(ctx context.Context, article domain.ArticleIdentifier, summary, content string, userId int64) (*url.URL, error) {
	//TODO: deal with variations in the capitalization of the article title.
	//TODO: check if user has permission to edit the wiki and the article in question.
	author, err := s.DB.GetUserURI(ctx, userId)
	if err != nil {
		return nil, err
	}

	articleId, ap, prev, err := s.DB.GetLastRevisionID(ctx, article)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			ap, err = s.CreateArticle(ctx, article, summary, content, userId)
		}
		return ap, err
	}
	// Deal with the case in which the article is remote.
	uri, err := s.DB.UpdateArticle(ctx, prev, articleId, userId, summary, content, nil)
	if err != nil {
		return nil, err
	}

	err = s.gateway.UpdateLocalArticle(ctx, uri, author, summary, articleId)

	return ap, err

}

func (s *AppService) GetArticle(ctx context.Context, title, author, host string) (article domain.ArticleFed, err error) {
	title = RemoveDuplicateSpaces(strings.ToLower(title))
	err = validate.Title(title)
	if err != nil {
		return
	}

	authorSql := sql.NullString{
		Valid: true,
		String: author,
	}

	article, err = s.DB.GetArticle(ctx, title, sql.NullString{
		Valid: host != "",
		String: host,
	}, authorSql)

	return
}

func (s *AppService) CreateArticle(ctx context.Context, articleId domain.ArticleIdentifier, summary, content string, userId int64) (*url.URL, error) {
	// TODO: validate title.
	user, err := s.DB.GetUserURI(ctx, userId)
	if err != nil {
		return nil, err
	}
	articleId.Title = RemoveDuplicateSpaces(articleId.Title)
	summary = RemoveDuplicateSpaces(summary)

	err = validate.Title(articleId.Title)
	if err != nil {
		return nil, err
	}

	// Handle remote article creation.
	apId := s.Config.Url.JoinPath("a", articleId.Title)
	article := domain.ArticleFed{
		ArticleCore: domain.ArticleCore{
			Title:     articleId.Title,
			Author: s.Config.Name,
			Host: s.Config.Domain,
			Content:   content,
			Language:  s.Config.Language,
			MediaType: s.Config.MediaType,
			Published: time.Now(),
		},
		ApID: apId,
		To: []*url.URL{
			domain.Public,
			s.Config.Url,
		},
		AttributedTo: s.Config.Url,
		Url:          apId,
	}

	diffs := s.FindDiff("", content)
	revision := domain.Revision{
		Summary:  summary,
		Diff:     diffs,
		Reviewed: false,
	}

	if err = s.DB.CreateLocalArticle(ctx, userId, article, revision); err != nil {
		return nil, err
	}

	err = s.gateway.CreateLocalArticle(ctx, article, user, summary)
	return apId, err
}

func (s *AppService) GetRevisionList(ctx context.Context, title, author, host string) ([]domain.Revision, error) {
	edits, err := s.DB.GetRevisionList(ctx, title, author, host)
	return edits, err
}

func (s *AppService) FindDiff(text1, text2 string) string {
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
