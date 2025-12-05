package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
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
	case streams.ActivityStreamsUpdateName:
		update, ok := asType.(vocab.ActivityStreamsUpdate)
		if !ok {
			return fmt.Errorf("failed to convert update activity")
		}
		return g.processUpdate(ctx, update)
	default:
		return fmt.Errorf("%w: %s", errors.ErrUnsupported, asType.GetTypeName())
	}
}

func (g *FedGatewayImpl) processUpdate(ctx context.Context, update vocab.ActivityStreamsUpdate) error {
	id, err := g.processId(update.GetJSONLDId())
	if err != nil {
		return err
	}

	actor, err := g.processActor(update.GetActivityStreamsActor())
	if err != nil {
		return err
	}

	exists, err := g.db.Exists(ctx, actor)
	if err != nil {
		return err
	}
	if !exists {
		payload, err := streams.Serialize(update)
		if err != nil {
			return err
		}
		task := Task{
			Type: Fetch,
			To: actor.String(),
			Next: &Task{
				Type: Process,
				Payload: payload,
			},
		}
		_, err = g.queue.Add(task).Save()
		return err
	}
	
	objProp := update.GetActivityStreamsObject()
	if objProp == nil || objProp.Len() == 0 {
		return fmt.Errorf("%w: object", federation.ErrMissingProperty)
	}

	var summary string
	if summaryProp := update.GetActivityStreamsSummary(); summaryProp != nil && summaryProp.Len() != 0 {
		summary = summaryProp.Begin().GetXMLSchemaString()
	}
	
	obj := objProp.Begin()
	if obj.IsIRI() {
		contentProp := update.GetActivityStreamsContent()
		if contentProp == nil || contentProp.Len() == 0 {
			return g.Fetch(obj.GetIRI())
		}

		//TODO: handle case when property is not an XMLSchemaString
		content := contentProp.Begin().GetXMLSchemaString()


		return g.db.UpdateFedArticle(ctx, obj.GetIRI(), id, actor, content, summary)
	}

	if t := obj.GetType(); t != nil {
		return g.processUpdatedObject(ctx, t, actor, id, summary)
	}

	return fmt.Errorf("%w: updated object", federation.ErrUnprocessablePropValue)
}

// processUpdatedObject handles the object of an update activity, which typically will be an article or note.
// Updating an article requires some extra logic that is not covered in the ProcessObject method, such as
// inserting a revision having the update's IRI.
func (g *FedGatewayImpl) processUpdatedObject(ctx context.Context, obj vocab.Type, actorIRI, updateIRI *url.URL, summary string) error {
	switch obj.GetTypeName() {
	case streams.ActivityStreamsArticleName:
		article, err := conversions.ConvertArticle(obj)
		if err != nil {
			return err
		}

		exists, err := g.db.Exists(ctx, article.ApID)
		if err != nil {
			return err
		}
		if !exists {
			raw, err := conversions.SerializeActivity(obj)
			if err != nil {
				return err
			}
			return g.db.PersistRemoteArticle(ctx, article, raw)
		}
		
		return g.db.UpdateFedArticle(ctx, article.ApID, updateIRI, actorIRI, article.Content, summary)
	default:
		return fmt.Errorf("%w: %s", federation.ErrUnsupported, obj.GetTypeName())
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