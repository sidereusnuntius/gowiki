package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (g *FedGatewayImpl) ProcessObject(ctx context.Context, asType vocab.Type) error {
	b, _ := streams.Serialize(asType)
	fmt.Printf("%+v\n", b)
	switch asType.GetTypeName() {
	case streams.ActivityStreamsArticleName:
		article, err := conversions.ConvertArticle(asType)
		if err != nil {
			return fmt.Errorf("invalid AS type")
		}

		raw, err := conversions.SerializeActivity(asType)
		if err != nil {
			return fmt.Errorf("serialization error: %w", err)
		}

		return g.db.PersistRemoteArticle(ctx, article, raw)
	case streams.ActivityStreamsFollowName:
		follow, ok := asType.(vocab.ActivityStreamsFollow)
		if !ok {
			return fmt.Errorf("failed to convert follow activity")
		}
		return g.processFollow(ctx, follow)
	default:
		return fmt.Errorf("%w: %s", errors.ErrUnsupported, asType.GetTypeName())
	}
}

func (g *FedGatewayImpl) processFollow(ctx context.Context, follow vocab.ActivityStreamsFollow) error {
	id, err := g.processId(follow.GetJSONLDId())
	if err != nil {
		return err
	}

	actor, err := g.processActor(follow.GetActivityStreamsActor())
	if err != nil {
		return err
	}

	obj, err := g.processObjectProperty(ctx, follow.GetActivityStreamsObject())
	if err != nil {
		return err
	}
	
	props, err := streams.Serialize(follow)
	if err != nil {
		return err
	}

	rawJSON, err := json.Marshal(props)
	if err != nil {
		return err
	}

	fmt.Printf("Actor: %s\nObject: %s\n", actor, obj)
	returnedId, err := g.db.Follow(ctx, domain.Follow{
		IRI:           id,
		Follower:      actor,
		Followee:      obj,
		FollowerInbox: nil,
		Raw:           rawJSON,
	})
	if err != nil {
		return err
	}

	// TODO: Ibis and Mastodon repeat the follow activity in the accept's object property.
	acceptId := g.cfg.Url.JoinPath("accept", strconv.Itoa(int(returnedId)))
	accept := conversions.NewAccept(acceptId, obj, id)

	return g.Deliver(ctx, accept, actor, obj)
}