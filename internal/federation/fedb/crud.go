package fedb

import (
	"context"
	"encoding/json"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/conversions"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (fd *FedDB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {
	obj, err := fd.DB.GetApObject(ctx, id)
	if err != nil {
		return
	}
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
	return nil
}

func (fd *FedDB) Update(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (fd *FedDB) Delete(ctx context.Context, id *url.URL) error {
	return nil
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