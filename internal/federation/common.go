package federation

import (
	"context"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams/vocab"
)

func (f *ApService) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return
}

func (f *ApService) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return
}

func (f *ApService) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return nil, nil
}

func (f *ApService) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	return
}