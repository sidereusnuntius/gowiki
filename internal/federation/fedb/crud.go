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
	id.Scheme = "http"
	log.Debug().Str("id", id.String()).Msg("trying to get ActivityStreams object")
	obj, err := fd.DB.GetApObject(ctx, id)
	if err != nil {
		return
	}
	log.Debug().Any("returned", obj).Msg("at Get:")

	if obj.RawJSON == "" {
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

func (fd *FedDB) Create(ctx context.Context, asType vocab.Type) error {
	log.Info().Any("obj", asType).Msg("creating AS object")
	props, err := streams.Serialize(asType)
	if err != nil {
		return err
	}

	rawJSON, err := json.Marshal(props)
	if err != nil {
		return err
	}
	return fd.DB.CreateApObject(ctx, domain.FedObj{
		Iri:     asType.GetJSONLDId().GetIRI(),
		RawJSON: string(rawJSON),
		ApType:  asType.GetTypeName(),
	}, 0)
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

	return fd.DB.UpdateAp(ctx, iri, string(bytes))
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
		t = conversions.ArticleToObject(a)
	}

	return
}
