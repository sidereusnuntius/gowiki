package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/client"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type FedGateway interface {
	Fetch(iri *url.URL) error
	Deliver(ctx context.Context, activity vocab.Type, to *url.URL, from *url.URL) error

	ProcessObject(ctx context.Context, asType vocab.Type) error
	ProcessOutbox(ctx context.Context, asType vocab.Type) error

	// Perhaps move these to a Notifier interface?
	CreateLocalArticle(ctx context.Context, article domain.ArticleFed, authorId *url.URL, summary string) error
	UpdateLocalArticle(ctx context.Context, updateURI, author *url.URL, summary string, id int64) error
	FollowRemoteActor(ctx context.Context, follower, followee *url.URL) error
}

type FedGatewayImpl struct {
	client *client.HttpClient
	db     db.DB
	queue  *backlite.Client
	cfg    *config.Configuration
}

func New(ctx context.Context, db db.DB, client *client.HttpClient, cfg *config.Configuration, blClient *backlite.Client) FedGateway {

	q := &FedGatewayImpl{
		db:     db,
		queue:  blClient,
		client: client,
		cfg:    cfg,
	}
	queue := backlite.NewQueue(q.processTask())
	q.queue.Register(queue)
	q.queue.Start(ctx)
	log.Info().Msg("started task queue")
	return q
}

func (q *FedGatewayImpl) Fetch(iri *url.URL) error {
	log.Debug().Str("iri", iri.String()).Msg("enqueing fetch task")
	task := Task{
		Type: Fetch,
		To:   iri.String(),
	}

	_, err := q.queue.Add(task).Save()
	return err
}

func (q *FedGatewayImpl) serializeAndPersist(ctx context.Context, activity vocab.Type, sender *url.URL) (map[string]any, error) {
	data, err := streams.Serialize(activity)
	if err != nil {
		log.Error().Err(err).Msg("activity serialization error")
		return nil, err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = q.db.AddOutbox(ctx, activity.GetTypeName(), bytes, activity.GetJSONLDId().GetIRI(), sender.JoinPath("outbox"))
	return data, err
}

func (q *FedGatewayImpl) BatchDeliver(ctx context.Context, activity vocab.Type, receivers []*url.URL, from *url.URL) error {
	data, err := q.serializeAndPersist(ctx, activity, from)
	if err != nil {
		return err
	}

	for _, to := range receivers {
		if err = q.rawDeliver(ctx, data, to, from); err != nil {
			log.Error().Err(err).Msg("delivery task enqueue")
		}
	}

	return nil
}

func (q *FedGatewayImpl) Deliver(ctx context.Context, activity vocab.Type, to *url.URL, from *url.URL) error {
	data, err := q.serializeAndPersist(ctx, activity, from)
	if err != nil {
		return err
	}

	return q.rawDeliver(ctx, data, to, from)
}

func (q *FedGatewayImpl) rawDeliver(ctx context.Context, activity map[string]any, to *url.URL, from *url.URL) error {
	var task = Task{
		Type:    Deliver,
		To:      to.String(),
		From:    from.String(),
		Payload: activity,
	}

	_, err := q.db.GetActorInbox(ctx, to)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			_, err = q.queue.Add(Task{
				Type: Fetch,
				To:   to.String(),
				Next: &task,
			}).Save()
		}
		return err
	}

	_, err = q.queue.Add(task).Save()
	if err != nil {
		log.Error().Err(err).Msg("adding delivery task to queue")
	}
	return err
}
