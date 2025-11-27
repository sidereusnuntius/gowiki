package queue

import (
	"context"
	"net/url"

	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/client"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

type ApQueue interface {
	Fetch(iri *url.URL) error
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