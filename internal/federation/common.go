package federation

import (
	"context"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams/vocab"
)

func (f *FedProto) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return
}

func (f *FedProto) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return
}

func (f *FedProto) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return nil, nil
}

func (f *FedProto) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	return
}