package fedb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/domain"
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
	return fd.DB.Follow(ctx, domain.Follow{
		IRI: id,
		Follower: actor,
		Followee: obj,
		FollowerInbox: nil,
		Raw: string(rawJSON),
	})
}

func (fd *FedDB) handleId(prop vocab.JSONLDIdProperty) (iri *url.URL, err error) {
	if prop == nil {
		return nil, fmt.Errorf("%w: id", ErrMissingProperty)
	}
	
	iri = prop.GetIRI()
	if iri == nil {
		err = fmt.Errorf("%w: id", ErrMissingProperty)
	}
	return
}

func (fd *FedDB) handleActorProp(prop vocab.ActivityStreamsActorProperty) (*url.URL, error){
	if prop == nil || prop.Len() == 0 {
		return nil, fmt.Errorf("%w: actor", ErrMissingProperty)
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
		return nil, fmt.Errorf("%w: object", ErrMissingProperty)
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
		return nil, fmt.Errorf("%w: %s", ErrNotFoundIRI, iri)
	}

	return iri, nil
}