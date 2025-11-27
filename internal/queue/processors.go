package queue

import (
	"context"
	"errors"
	"net/url"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
)

func (q *apQueueImpl) register() {
	fetchQueue := backlite.NewQueue[FetchJob](q.fetch())

	q.queues.Register(fetchQueue)
}

func (q *apQueueImpl) fetch() func(context.Context, FetchJob) error {
	return func(ctx context.Context, task FetchJob) error {
		iri, err := url.Parse(task.Iri)
		if err != nil {
			return err
		}
		defer func(){
			if err != nil {
				log.Error().Err(err).Msg("fetch failed")
			}
		}()
	
		fetchedAt := time.Now()
		asType, err := q.client.Get(ctx, iri)
		if err != nil {
			return err
		}
	
		switch asType.GetTypeName() {
		case streams.ActivityStreamsPersonName:
			person, _ := asType.(vocab.ActivityStreamsPerson)
			u, err := conversions.ActorToUser(person)
			if err != nil {
				return err
			}
			return q.db.InsertOrUpdateUser(ctx, u, fetchedAt)
		default:
			return errors.New("unprocessable entity")
		}
	}
}