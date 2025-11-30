package queue

import (
	"context"
	"errors"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (q *apQueueImpl) CreateLocalArticle(ctx context.Context, article domain.ArticleFed, authorId *url.URL, summary string) error {
	id := article.ApID.JoinPath("create")
	a := article.CreateAP(id, authorId, q.cfg.Url, summary)

	followers, err := q.db.GetFollowers(ctx, q.cfg.Url)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			err = nil
		}
		return err
	}

	for _, f := range followers {
		if err = q.Deliver(ctx, a, f, q.cfg.Url); err != nil {
			log.Error().Err(err).Str("to", f.String()).Msg("failed to enqueue delivery job")
		}
	}

	return nil
}