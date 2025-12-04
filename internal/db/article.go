package db

import (
	"context"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Article interface {
	GetRevisionList(ctx context.Context, title string) ([]domain.Revision, error)
	GetLocalArticle(ctx context.Context, title string) (domain.ArticleCore, error)
	// UpdateArticle creates a new revision and updates the article. apId is the IRI of the revision; it must only
	// be set when the article has been updated by an activity received from a remote host, and that activity
	// will have its own id. For local edits, apId is nil, and a new IRI is generated.
	UpdateArticle(ctx context.Context, prevId, articleId, userId int64, summary, newContent string, apId *url.URL) (URI *url.URL, err error)
	GetLastRevisionID(ctx context.Context, title string) (int64, *url.URL, int64, error)
	CreateLocalArticle(ctx context.Context, userId int64, article domain.ArticleFed, initialEdit domain.Revision) (err error)
	GetArticleById(ctx context.Context, id int64) (domain.ArticleFed, error)
}
