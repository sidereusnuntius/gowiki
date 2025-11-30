package core

import (
	"context"
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
func (s *AppService) AlterArticle(ctx context.Context, title, summary, content string, userId int64) (*url.URL, error) {
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

func (s *AppService) GetLocalArticle(ctx context.Context, title string) (article domain.ArticleCore, err error) {
	title = RemoveDuplicateSpaces(title)
	err = validate.Title(title)
	if err != nil {
		return
	}

	article, err = s.DB.GetLocalArticle(ctx, title)
	return
}

func (s *AppService) CreateArticle(ctx context.Context, title, summary, content string, userId int64) (*url.URL, error) {
	// TODO: validate title.
	user, err := s.DB.GetUserURI(ctx, userId)
	if err != nil {
		return nil, err
	}
	title = RemoveDuplicateSpaces(title)
	summary = RemoveDuplicateSpaces(summary)

	err = validate.Title(title)
	if err != nil {
		return nil, err
	}


	apId := s.Config.Url.JoinPath("a", title)
	article := domain.ArticleFed{
		ArticleCore: domain.ArticleCore{
			Title:     title,
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
		AttributedTo: user,
		Url:  apId,
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

	err = s.fedgateway.CreateLocalArticle(ctx, article, user, summary)
	return apId, err
}

func (s *AppService) GetRevisionList(ctx context.Context, title string) ([]domain.Revision, error) {
	edits, err := s.DB.GetRevisionList(ctx, title)
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
