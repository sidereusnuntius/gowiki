package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (g *FedGatewayImpl) FollowRemoteActor(ctx context.Context, follower, followee *url.URL) error {
	inbox, err := g.db.GetActorInbox(ctx, followee)
	if err != nil && !errors.Is(err, db.ErrNotFound){
		return err
	}

	_, followIRI, err := g.db.Follow(ctx, domain.Follow{
		Follower: follower,
		Followee: followee,
		FollowerInbox: inbox,
	})
	if err != nil {
		return err
	}

	follow := conversions.NewFollow(followIRI, follower, followee)
	followMap, err := streams.Serialize(follow)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(followMap)
	if err != nil {
		return err
	}

	if err = g.db.UpdateAp(ctx, followIRI, raw); err != nil {
		return err
	}

	return g.Deliver(ctx, follow, followee, follower)
}

func (q *FedGatewayImpl) UpdateLocalArticle(ctx context.Context, updateURI, author *url.URL, summary string, articleId int64) error {
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

func (q *FedGatewayImpl) CreateLocalArticle(ctx context.Context, article domain.ArticleFed, authorId *url.URL, summary string) error {
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
