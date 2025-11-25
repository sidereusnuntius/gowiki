package fedb

import (
	"context"
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
)

func (fd *FedDB) Followers(c context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	return
}

func (fd *FedDB) Following(c context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	return
}

func (fd *FedDB) Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error) {
	return
}
