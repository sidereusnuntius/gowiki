package fedb

import (
	"context"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/rs/zerolog/log"
)

func (fd *FedDB) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return
}

func (fd *FedDB) Following(ctx context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	followingIRIs, err := fd.DB.GetFollowing(ctx, actorIRI)
	log.Debug().Msg("at Following()")
	if err != nil {
		return
	}

	following = streams.NewActivityStreamsCollection()
	id := streams.NewJSONLDIdProperty()
	id.Set(actorIRI.JoinPath("following"))
	following.SetJSONLDId(id)

	items := streams.NewActivityStreamsItemsProperty()
	for _, iri := range followingIRIs {
		items.AppendIRI(iri)
	}
	following.SetActivityStreamsItems(items)

	total := streams.NewActivityStreamsTotalItemsProperty()
	total.Set(items.Len())
	following.SetActivityStreamsTotalItems(total)
	return
}

func (fd *FedDB) Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error) {
	return
}
