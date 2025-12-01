package queue

import (
	"context"
	"errors"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (q *apQueueImpl) UpdateLocalArticle(ctx context.Context, updateURI, author *url.URL, summary string, articleId int64) error {
	article, err := q.db.GetArticleById(ctx, articleId)
	if err != nil {
		return err
	}

	followers, err := q.db.GetFollowers(ctx, q.cfg.Url)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			err = nil
		}
		return err
	}
	
	update := article.UpdateAP(updateURI, author, q.cfg.Url, summary)
	return q.BatchDeliver(ctx, update, followers, author)
}

func (q *apQueueImpl) CreateLocalArticle(ctx context.Context, article domain.ArticleFed, authorId *url.URL, summary string) error {
	id := article.ApID.JoinPath("create")
	create := article.CreateAP(id, authorId, q.cfg.Url, summary)

	followers, err := q.db.GetFollowers(ctx, q.cfg.Url)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			err = nil
		}
		return err
	}

	return q.BatchDeliver(ctx, create, followers, q.cfg.Url)
}