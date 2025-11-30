package fedb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

// This file contains methods to handle the different types of activities supported by the wiki.

func (fd *FedDB) handleFollow(ctx context.Context, follow vocab.ActivityStreamsFollow) error {
	id, err := fd.handleId(follow.GetJSONLDId())
	if err != nil {
		return err
	}

	actor, err := fd.handleActorProp(follow.GetActivityStreamsActor())
	if err != nil {

		return err
	}

	obj, err := fd.handleObjProp(ctx, follow.GetActivityStreamsObject())
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
	returnedId, err := fd.DB.Follow(ctx, domain.Follow{
		IRI: id,
		Follower: actor,
		Followee: obj,
		FollowerInbox: nil,
		Raw: string(rawJSON),
	})
	if err != nil {
		return err
	}


	// TODO: Ibis and Mastodon repeat the follow activity in the accept's object property.
	acceptId := fd.Config.Url.JoinPath("accept", strconv.Itoa(int(returnedId)))
	accept := conversions.NewAccept(acceptId, obj, id)
	
	if err = fd.Queue.Deliver(ctx, accept, actor, obj); err != nil {
		return err
	}

	return fd.Create(ctx, accept)
}

func (fd *FedDB) handleId(prop vocab.JSONLDIdProperty) (iri *url.URL, err error) {
	if prop == nil {
		return nil, fmt.Errorf("%w: id", federation.ErrMissingProperty)
	}
	
	iri = prop.GetIRI()
	if iri == nil {
		err = fmt.Errorf("%w: id", federation.ErrMissingProperty)
	}
	return
}

func (fd *FedDB) handleActorProp(prop vocab.ActivityStreamsActorProperty) (*url.URL, error){
	if prop == nil || prop.Len() == 0 {
		return nil, fmt.Errorf("%w: actor", federation.ErrMissingProperty)
	}
	
	actor := prop.At(0)

	// Ensure both the sending instance and the actor are stored in the database.
	if actor.IsIRI() {
		return actor.GetIRI(), nil
	} else if actor.IsActivityStreamsPerson() {
		// TODO!
	}
	panic("not implemented yet")
}

func (fd *FedDB) handleObjProp(ctx context.Context, prop vocab.ActivityStreamsObjectProperty) (*url.URL, error) {
	if prop == nil || prop.Len() == 0 {
		return nil, fmt.Errorf("%w: object", federation.ErrMissingProperty)
	}

	var iri *url.URL
	obj := prop.At(0)
	if obj.IsIRI() {
		iri = obj.GetIRI()
	}

	exists, err := fd.Exists(ctx, iri)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("%w: %s", federation.ErrNotFoundIRI, iri)
	}

	return iri, nil
}