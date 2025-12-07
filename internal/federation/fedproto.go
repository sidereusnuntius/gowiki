package federation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

var (
	ErrMissingProperty        = errors.New("missing property")
	ErrUnprocessablePropValue = errors.New("unprocessable")
	ErrNotFoundIRI            = errors.New("unknown IRI")
	ErrUnsupported            = errors.New("unsupported")
	ErrConflict               = errors.New("conflict")
	ErrForbidden              = errors.New("forbidden")
)

type Verifier interface {
	Verify(ctx context.Context, r *http.Request) error
}

type ApService struct {
	DB       db.DB
	Verifier Verifier
}

func (f ApService) GetCollection(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	iri := r.URL
	size, start, err := f.DB.GetCollectionStart(ctx, iri)
	if err != nil {
		return err
	}

	collection := streams.NewActivityStreamsOrderedCollection()
	id := streams.NewJSONLDIdProperty()
	id.SetIRI(iri)
	collection.SetJSONLDId(id)

	total := streams.NewActivityStreamsTotalItemsProperty()
	total.Set(int(size))
	collection.SetActivityStreamsTotalItems(total)

	query, err := url.Parse("?last=" + strconv.FormatInt(start, 10))
	if err != nil {
		return err
	}

	first := streams.NewActivityStreamsFirstProperty()
	first.SetIRI(iri.ResolveReference(query))
	collection.SetActivityStreamsFirst(first)

	props, err := streams.Serialize(collection)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/activity+json")
	return encoder.Encode(props)
}

// AuthenticatePostInbox implements pub.FederatingProtocol.
func (f ApService) AuthenticatePostInbox(ctx context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	out = ctx
	log.Debug().Msg("at AuthenticatePostInbox()")
	if err = f.Verifier.Verify(ctx, r); err != nil {
		return
	}
	authenticated = true
	return
}

// Blocked implements pub.FederatingProtocol.
func (f ApService) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	log.Debug().Msg("at Blocked():")
	return
}

// DefaultCallback implements pub.FederatingProtocol.
func (f ApService) DefaultCallback(c context.Context, activity pub.Activity) error {
	log.Debug().Msg("at DefautlCallbacks():")
	return nil
}

// FederatingCallbacks implements pub.FederatingProtocol.
func (f ApService) FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []any, err error) {
	log.Debug().Msg("at AuthenticatePostInbox()")
	wrapped = pub.FederatingWrappedCallbacks{
		Follow: func(ctx context.Context, asf vocab.ActivityStreamsFollow) error {
			log.Info().Msg("Received a follow activity")
			return nil
		},
		Accept: func(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
			log.Info().Msg("received an accept activity")
			return nil
		},
	}
	return
}

// FilterForwarding implements pub.FederatingProtocol.
func (f ApService) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	log.Debug().Msg("at FilterForwarding():")
	return
}

// GetInbox implements pub.FederatingProtocol.
func (f ApService) GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	log.Debug().Msg("at GetInbox():")

	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}

// MaxDeliveryRecursionDepth implements pub.FederatingProtocol.
func (f ApService) MaxDeliveryRecursionDepth(c context.Context) int {
	log.Debug().Msg("at MaxDeliveryRecursionDepth():")
	return 7
}

// MaxInboxForwardingRecursionDepth implements pub.FederatingProtocol.
func (f ApService) MaxInboxForwardingRecursionDepth(c context.Context) int {
	log.Debug().Msg("at MaxInboxForwardingRecursionDepth():")
	return 2
}

// PostInboxRequestBodyHook implements pub.FederatingProtocol.
func (f ApService) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	log.Debug().Msg("at PostInboxRequestBodyHook()")
	return c, nil
}
