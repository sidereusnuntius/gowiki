package queue

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (q *apQueueImpl) processTask() func(context.Context, Task) error {
	return func(ctx context.Context, task Task) error {
		var to, from *url.URL
		to, err := url.Parse(task.To)
		if err != nil {
			return err
		}

		switch task.Type {
		case Fetch:
			err = q.fetch(ctx, to)
		case Deliver:
			if from, err = url.Parse(task.From); err != nil {
				return err
			}
			err = q.deliver(ctx, to, from, task.Payload)
		default:
			err = errors.New("unsupported task type")
		}

		if err != nil {
			return err
		}

		if task.Next != nil {
			_, err = backlite.FromContext(ctx).Add(*task.Next).Save()
		}
		return err
	}
}

func (q *apQueueImpl) fetch(ctx context.Context, iri *url.URL) error {
	log.Debug().Str("iri", iri.String()).Msg("fetching IRI")

	fetchedAt := time.Now()
	asType, err := q.client.Get(ctx, iri)
	if err != nil {
		log.Error().Err(err).Msg("fetch error")
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
		return q.db.InsertOrUpdateUser(ctx, u, fetchedAt)
	default:
		return errors.New("unprocessable entity")
	}
}

func (q *apQueueImpl) deliver(ctx context.Context, to, from *url.URL, payload map[string]any) error {
	inbox, err := q.db.GetActorInbox(ctx, to)
	if err != nil {
		return fmt.Errorf("%w: actor's inbox not found (actor = %s)", err, to)
	}

	log.Debug().Str("to", to.String()).
		Str("inbox", inbox.String()).
		Msg("delivering activity")

	// Move inbox resolve to client.
	if err = q.client.DeliverAs(ctx, payload, inbox, from); err != nil {
		err = fmt.Errorf("delivery error: %w", err)
	}
	return err
}
