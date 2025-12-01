package fedb

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"codeberg.org/gruf/go-mutexes"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/queue"
)

type FedDB struct {
	DB     db.DB
	Queue  queue.ApQueue
	Config config.Configuration
	locks  *mutexes.MutexMap
}

// GetOutbox implements pub.Database.
func (fd *FedDB) GetOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	panic("unimplemented")
}

// InboxesForIRI implements pub.Database.
func (fd *FedDB) InboxesForIRI(c context.Context, iri *url.URL) (inboxIRIs []*url.URL, err error) {
	panic("unimplemented")
}

// SetOutbox implements pub.Database.
func (fd *FedDB) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	panic("unimplemented")
}

func New(DB db.DB, queue queue.ApQueue, config config.Configuration) FedDB {
	locks := mutexes.MutexMap{}
	return FedDB{
		DB:     DB,
		Config: config,
		locks:  &locks,
		Queue:  queue,
	}
}

func (fd *FedDB) Lock(c context.Context, id *url.URL) (unlock func(), err error) {
	unlockFunc := fd.locks.Lock(id.String())

	if unlockFunc == nil {
		err = errors.New("lock failed")
		return
	}
	unlock = func() {
		unlockFunc()
	}
	return
}

func (fd *FedDB) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	log.Debug().Msg("at InboxContains()")
	contains, err = fd.DB.CollectionContains(c, inbox, id)
	log.Debug().Bool("contains", contains).Send()
	return
}

func (fd *FedDB) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	log.Debug().Msg("at GetInbox()")
	inbox = streams.NewActivityStreamsOrderedCollectionPage()
	return
}

func (fd *FedDB) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	log.Debug().Msg("at SetInbox()")
	iter := inbox.GetActivityStreamsOrderedItems()

	for item := iter.Begin(); item != nil; item = item.Next() {
		if item.IsIRI() {
			fmt.Printf("IRI: %s\n", item.GetIRI())
		}
	}

	// Most certainly won't be supported.
	return nil
}

func (fd *FedDB) Owns(ctx context.Context, id *url.URL) (owns bool, err error) {
	log.Debug().Msg("at Owns()")
	// TODO: rename Domain to Host
	if id.Host != fd.Config.Domain {
		return
	}
	owns, err = fd.Exists(ctx, id)
	return
}

func (fd *FedDB) Exists(ctx context.Context, id *url.URL) (exists bool, err error) {
	log.Debug().Msg("at Exists(): checking if " + id.String() + " exists")
	exists, err = fd.DB.Exists(ctx, id)
	return
}

func (fd *FedDB) ActorForOutbox(ctx context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	log.Debug().Msg("at ActorForOutbox()")
	actorIRI, err = fd.DB.ActorIdByOutbox(ctx, outboxIRI)
	return
}

func (fd *FedDB) ActorForInbox(ctx context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	log.Debug().Msg("at ActorForInbox")
	actorIRI, err = fd.DB.ActorIdByInbox(ctx, inboxIRI)
	if err != nil {
		log.Error().Str("inbox IRI", inboxIRI.String()).Err(err).Send()
	}
	return
}

func (fd *FedDB) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	log.Debug().Msg("at OutboxForInbox()")
	outboxIRI, err = fd.DB.OutboxForInbox(ctx, inboxIRI)
	return
}

func (fd *FedDB) NewID(ctx context.Context, t vocab.Type) (id *url.URL, err error) {
	log.Debug().Msg("at NewID()")
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
