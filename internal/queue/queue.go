package queue

import (
	"context"
	"errors"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/client"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

type ApQueue interface {
	Fetch(iri *url.URL) error
	Deliver(ctx context.Context, activity vocab.Type, to *url.URL, from *url.URL) error
}

type apQueueImpl struct {
	db db.DB
	queues *backlite.Client
	client *client.HttpClient
}

func New(ctx context.Context, db db.DB, client *client.HttpClient, blClient *backlite.Client) ApQueue {
	
	q := &apQueueImpl{
		db: db,
		queues: blClient,
		client: client,
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

func (q *apQueueImpl) Deliver(ctx context.Context, activity vocab.Type, to *url.URL, from *url.URL) error {
	_, err := q.db.GetActorInbox(ctx, to)
	var inboxUnknown bool
	if err != nil {
		log.Error().Err(err).Msg("at delivery")
		if !errors.Is(err, db.ErrNotFound) {
			return err
		}
		inboxUnknown = true
	}

	data, err := streams.Serialize(activity)
	if err != nil {
		return err
	}

	var task backlite.Task = PostJob{
		To: to.String(),
		From: from.String(),
		Body: data,
	}

	if inboxUnknown {
		task = FetchJob{
			Iri: to.String(),
			Next: task,
		}
	}

	_, err = q.queues.Add(task).Save()
	return err
}