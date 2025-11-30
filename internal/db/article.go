package db

import (
	"context"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Article interface {
	GetRevisionList(ctx context.Context, title string) ([]domain.Revision, error)
	GetLocalArticle(ctx context.Context, title string) (domain.ArticleCore, error)
	UpdateArticle(ctx context.Context, prevId, articleId, userId int64, summary, newContent string) (URI *url.URL, err error)
	GetLastRevisionID(ctx context.Context, title string) (int64, *url.URL, int64, error)
	CreateLocalArticle(ctx context.Context, userId int64, article domain.ArticleFed, initialEdit domain.Revision) (err error)
	GetArticleById(ctx context.Context, id int64) (domain.ArticleFed, error)
}
