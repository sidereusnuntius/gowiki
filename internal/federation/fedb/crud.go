package fedb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (fd *FedDB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	log.Debug().Str("id", id.String()).Msg("at Get(): trying to get ActivityStreams object")
	obj, err := fd.DB.GetApObject(ctx, id)
	if err != nil {
		return
	}
	log.Debug().Any("returned", obj).Msg("at Get:")

	if obj.ApType == streams.ActivityStreamsOrderedCollectionName {
		value, err = fd.handleCollection(ctx, id)
	} else if obj.RawJSON == nil {
		value, err = fd.routeQuery(ctx, obj.LocalTable, obj.LocalId)
	} else {
		var temp map[string]any
		err = json.Unmarshal([]byte(obj.RawJSON), &temp)
		if err != nil {
			return
		}

		value, err = streams.ToType(ctx, temp)
	}

	return
}

func (fd *FedDB) Create(ctx context.Context, asType vocab.Type) (err error) {
	err = fd.Gateway.ProcessObject(ctx, asType)
	if err != nil {
		err = fmt.Errorf("creation error: %w", err)
	}
	return
}

func (fd *FedDB) Update(ctx context.Context, asType vocab.Type) error {
	iri := asType.GetJSONLDId().GetIRI()
	exists, err := fd.DB.Exists(ctx, iri)
	if err != nil {
		return err
	}

	if !exists {

	}

	s, err := streams.Serialize(asType)
	if err != nil {
		return fmt.Errorf("serializing object %s: %w", iri, err)
	}

	bytes, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("Update(): serialization error: %w", err)
	}

	return fd.DB.UpdateAp(ctx, iri, bytes)
}

func (fd *FedDB) Delete(ctx context.Context, id *url.URL) error {
	return fd.DB.DeleteAp(ctx, id)
}

func (fd *FedDB) routeQuery(ctx context.Context, table string, id int64) (t vocab.Type, err error) {
	switch table {
	case "users":
		var u domain.UserFed
		u, err = fd.DB.GetUserByID(ctx, id)
		if err != nil {
			return
		}
		t = conversions.UserToActor(u)
	case "articles":
		var a domain.ArticleFed
		a, err = fd.DB.GetArticleById(ctx, id)
		if err != nil {
			return
		}
		t = a.ConvertToAp()
	case "collectives":
		var c domain.Collective
		c, err = fd.DB.GetCollectiveById(ctx, id)
		if err != nil {
			return
		}

		t = conversions.GroupToActor(c)
	}

	return
}

func (fd *FedDB) handleCollection(ctx context.Context, id *url.URL) (t vocab.ActivityStreamsOrderedCollection, err error) {
	items, err := fd.DB.GetCollectionMemberIRIS(ctx, id)
	if err != nil {
		return
	}

	t = streams.NewActivityStreamsOrderedCollection()
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(id)
	t.SetJSONLDId(idProp)

	l := streams.NewActivityStreamsTotalItemsProperty()
	l.Set(len(items))
	t.SetActivityStreamsTotalItems(l)

	itemsProp := streams.NewActivityStreamsOrderedItemsProperty()
	for _, i := range items {
		itemsProp.AppendIRI(i)
	}

	t.SetActivityStreamsOrderedItems(itemsProp)
	return
}
