package federation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/rs/zerolog/log"
)

type ApService struct {	
}

// AuthenticatePostInbox implements pub.FederatingProtocol.
func (f ApService) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	log.Debug().Msg("at AuthenticatePostInbox()")
	out = c
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
func (f ApService) FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	log.Debug().Msg("at AuthenticatePostInbox()")
	wrapped = pub.FederatingWrappedCallbacks{
		Follow: func(ctx context.Context, asf vocab.ActivityStreamsFollow) error {
			log.Info().Msg("Received a follow activity")
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
	b, _ := streams.Serialize(activity)
	fmt.Printf("%+v\n", b)
	return c, nil
}
