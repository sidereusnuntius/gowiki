package queue

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

type ApQueue interface {
	Fetch(iri *url.URL) error
	Deliver(ctx context.Context, activity vocab.Type, to *url.URL, from *url.URL) error

	// Perhaps move these to a Notifier interface?
	CreateLocalArticle(ctx context.Context, article domain.ArticleFed, authorId *url.URL, summary string) error
	UpdateLocalArticle(ctx context.Context, updateURI, author *url.URL, summary string, id int64) error
}

type apQueueImpl struct {
	client *client.HttpClient
	db db.DB
	queues *backlite.Client
	cfg *config.Configuration
}

func New(ctx context.Context, db db.DB, client *client.HttpClient, cfg *config.Configuration, blClient *backlite.Client) ApQueue {
	
	q := &apQueueImpl{
		db: db,
		queues: blClient,
		client: client,
		cfg: cfg,
	}
	q.register()
	q.queues.Start(ctx)
	log.Info().Msg("started task queue")
	return q
}

func (q *apQueueImpl) Fetch(iri *url.URL) error {
	log.Debug().Str("iri", iri.String()).Msg("enqueing fetch task")
	task := FetchJob{
		Iri: iri.String(),
	}
	_, err := q.queues.Add(task).Save()
	return err
}

func (q *apQueueImpl) serializeAndPersist(ctx context.Context, activity vocab.Type, sender *url.URL) (map[string]any, error) {
	data, err := streams.Serialize(activity)
	if err != nil {
		log.Error().Err(err).Msg("activity serialization error")
		return nil, err
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	err = q.db.AddOutbox(ctx, string(bytes), activity.GetTypeName(), activity.GetJSONLDId().GetIRI(), sender.JoinPath("outbox"))
	return data, err
}

func (q *apQueueImpl) BatchDeliver(ctx context.Context, activity vocab.Type, receivers []*url.URL, from *url.URL) error {
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

func (q *apQueueImpl) Deliver(ctx context.Context, activity vocab.Type, to *url.URL, from *url.URL) error {
	data, err := q.serializeAndPersist(ctx, activity, from)
	if err != nil {
		return err
	}

	return q.rawDeliver(ctx, data, to, from)
}

func (q *apQueueImpl) rawDeliver(ctx context.Context, activity map[string]any, to *url.URL, from *url.URL) error {
	var task = PostJob{
		To: to.String(),
		From: from.String(),
		Body: activity,
	}

	_, err := q.db.GetActorInbox(ctx, to)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			_, err = q.queues.Add(FetchJob{
				Iri: to.String(),
				Next: &task,
			}).Save()
		}
		return err
	}

	_, err = q.queues.Add(task).Save()
	if err != nil {
		log.Error().Err(err).Msg("adding delivery task to queue")
	}
	return err
}