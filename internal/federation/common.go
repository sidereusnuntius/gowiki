package federation

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/rs/zerolog/log"
)

func (f *ApService) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	authenticated = true
	return
}

func (f *ApService) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	log.Debug().Msg("at AuthenticateGetOutbox()")
	out = c
	authenticated = true
	return
}

func (f *ApService) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	uri := r.URL
	var last int64
	t, err := strconv.ParseInt(uri.Query().Get("last"), 10, 64)
	if err == nil {
		last = t
	}

	query := uri.RawQuery
	uri.RawQuery = ""
	results, err := f.DB.GetCollectionActivities(c, uri, last)
	if err != nil {
		return nil, err
	}

	uri.RawQuery = query
	page := streams.NewActivityStreamsOrderedCollectionPage()
	id := streams.NewJSONLDIdProperty()
	id.SetIRI(uri)
	page.SetJSONLDId(id)

	items := streams.NewActivityStreamsOrderedItemsProperty()
	var temp vocab.Type
	for _, a := range results {
		temp, err = streams.ToType(c, a)
		if err != nil {
			return nil, err
		}

		if err = items.AppendType(temp); err != nil {
			return nil, err
		}
	}

	page.SetActivityStreamsOrderedItems(items)
	return page, nil
}

func (f *ApService) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	log.Debug().Msg("at NewTransport")
	return
}
