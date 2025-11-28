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
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (q *apQueueImpl) register() {
	fetchQueue := backlite.NewQueue[FetchJob](q.fetch())
	deliveryQueue := backlite.NewQueue[PostJob](q.deliver())

	q.queues.Register(fetchQueue)
	q.queues.Register(deliveryQueue)
}

func (q *apQueueImpl) fetch() func(context.Context, FetchJob) error {
	return func(ctx context.Context, task FetchJob) error {
		log.Debug().Str("iri", task.Iri).Msg("fetching IRI")
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
			var u domain.UserFed
			u, err = conversions.ActorToUser(person)
			if err != nil {
				return err
			}
			err = q.db.InsertOrUpdateUser(ctx, u, fetchedAt)
		default:
			err = errors.New("unprocessable entity")
		}

		if err != nil || task.Next == nil {
			return err
		}

		_, err = backlite.FromContext(ctx).Add(task.Next).Save()
		return err
	}
}

func (q *apQueueImpl) deliver() func(context.Context, PostJob) error {
	return func(ctx context.Context, pj PostJob) error {
		to, err := url.Parse(pj.To)

		if err != nil {
			return err
		}

		inbox, err := q.db.GetActorInbox(ctx, to)
		if err != nil {
			log.Error().Str("receiver", pj.To).Err(err).Msg("actor's inbox not found")
			return err
		}

		log.Debug().Str("to", pj.To).
			Str("inbox", inbox.String()).
			Msg("delivering activity")

		from, err := url.Parse(pj.From)
		if err != nil {
			return err
		}

		if err = q.client.DeliverAs(ctx, pj.Body, to, from); err != nil || pj.Next == nil {
			return err
		}

		_, err = backlite.FromContext(ctx).Add(pj.Next).Save()
		return err
	}
}