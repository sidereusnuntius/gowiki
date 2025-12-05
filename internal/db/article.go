package db

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Article interface {
	GetRevisionList(ctx context.Context, title, author, host string) ([]domain.Revision, error)
	GetArticle(ctx context.Context, title string, host, author sql.NullString) (domain.ArticleFed, error)
	// UpdateArticle creates a new revision and updates the article. apId is the IRI of the revision; it must only
	// be set when the article has been updated by an activity received from a remote host, and that activity
	// will have its own id. For local edits, apId is nil, and a new IRI is generated.
	UpdateArticle(ctx context.Context, prevId, articleId, userId int64, summary, newContent string, apId *url.URL) (URI *url.URL, err error)
	GetLastRevisionID(ctx context.Context, article domain.ArticleIdentifier) (articleID int64, articleIRI *url.URL, lastRevisionID int64, error error)
	CreateLocalArticle(ctx context.Context, userId int64, article domain.ArticleFed, initialEdit domain.Revision) (err error)
	GetArticleById(ctx context.Context, id int64) (domain.ArticleFed, error)
}
