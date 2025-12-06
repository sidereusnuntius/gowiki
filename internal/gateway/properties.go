package gateway

import (
	"context"
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func (g *FedGatewayImpl) processObjectProperty(ctx context.Context, prop vocab.ActivityStreamsObjectProperty) (*url.URL, error) {
	if prop == nil || prop.Len() == 0 {
		return nil, fmt.Errorf("%w: object", federation.ErrMissingProperty)
	}

	var iri *url.URL
	obj := prop.At(0)
	if obj.IsIRI() {
		iri = obj.GetIRI()
	}

	exists, err := g.db.Exists(ctx, iri)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("%w: %s", federation.ErrNotFoundIRI, iri)
	}

	return iri, nil
}

func (g *FedGatewayImpl) processId(prop vocab.JSONLDIdProperty) (iri *url.URL, err error) {
	if prop == nil {
		return nil, fmt.Errorf("%w: id", federation.ErrMissingProperty)
	}

	iri = prop.GetIRI()
	if iri == nil {
		err = fmt.Errorf("%w: id", federation.ErrMissingProperty)
	}
	return
}

func (g *FedGatewayImpl) processActor(prop vocab.ActivityStreamsActorProperty) (*url.URL, error) {
	if prop == nil || prop.Len() == 0 {
		return nil, fmt.Errorf("%w: actor", federation.ErrMissingProperty)
	}

	// TODO: maybe handle multiple actors.
	iter := prop.Begin()
	if iter.IsIRI() {
		return iter.GetIRI(), nil
	}

	t := iter.GetType()
	if t == nil {
		return nil, fmt.Errorf("%w: actor has an unsupported value", federation.ErrUnprocessablePropValue)
	}

	idProp := t.GetJSONLDId()
	if idProp == nil {
		return nil, fmt.Errorf("%w: actor's id", federation.ErrMissingProperty)
	}

	return idProp.Get(), nil
}
