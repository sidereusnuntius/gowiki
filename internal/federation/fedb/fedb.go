package fedb

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"codeberg.org/gruf/go-mutexes"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

type FedDB struct {
	DB     db.DB
	Config config.Configuration
	locks  mutexes.MutexMap
}

func New(DB db.DB, config config.Configuration) FedDB {
	return FedDB{
		DB: DB,
		Config: config,
		locks: mutexes.MutexMap{},
	}
}

func (fd *FedDB) Lock(c context.Context, id *url.URL) (unlock func(), err error) {
	unlock = fd.locks.Lock(id.String())
	return
}

func (fd *FedDB) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	//TODO
	return
}

func (fd *FedDB) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	//TODO
	return
}

func (fd *FedDB) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	// Most certainly won't be supported.
	return nil
}

func (fd *FedDB) Owns(ctx context.Context, id *url.URL) (owns bool, err error) {
	// TODO: rename Domain to Host
	if id.Host != fd.Config.Domain {
		return
	}
	owns, err = fd.Exists(ctx, id)
	return
}

func (fd *FedDB) Exists(ctx context.Context, id *url.URL) (exists bool, err error) {
	exists, err = fd.DB.Exists(ctx, id)
	return
}

func (fd *FedDB) ActorForOutbox(ctx context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	actorIRI, err = fd.DB.ActorIdByOutbox(ctx, outboxIRI)
	return
}

func (fd *FedDB) ActorForInbox(ctx context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	actorIRI, err = fd.DB.ActorIdByInbox(ctx, inboxIRI)
	return
}

func (fd *FedDB) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	outboxIRI, err = fd.DB.OutboxForInbox(ctx, inboxIRI)
	return
}

func (fd *FedDB) NewID(ctx context.Context, t vocab.Type) (id *url.URL, err error) {
	var title, path string
	switch v := t.(type) {
	case vocab.ActivityStreamsArticle:
		name := v.GetActivityStreamsName()
		if name == nil {
			return nil, errors.New("name property not present")
		}
		title = name.Begin().GetXMLSchemaString()
		path = "a"
	default:
		return nil, fmt.Errorf("unsupported type: %s", v)
	}

	if title == "" {
		return nil, errors.New("empty title")
	}

	id = fd.Config.Url.JoinPath(path, title)
	return
}
