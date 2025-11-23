package db

import (
	"context"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Fed interface {
	ActorIdByInbox(ctx context.Context, iri *url.URL) (*url.URL, error)
	ActorIdByOutbox(ctx context.Context, iri *url.URL) (*url.URL, error)
	OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (*url.URL, error)
	GetUserFed(ctx context.Context, id *url.URL) (user domain.UserFed, err error)
	GetInstanceIdOrCreate(ctx context.Context, hostname string) (id int64, err error)
	GetApObject(ctx context.Context, iri *url.URL) (domain.FedObj, error)
	CreateApObject(ctx context.Context, obj domain.FedObj, fetched int64) error
	GetUserByID(ctx context.Context, id int64) (domain.UserFed, error)
	Exists(ctx context.Context, id *url.URL) (bool, error)
	UpdateAp(ctx context.Context, id *url.URL, rawJSON string) (error)
	DeleteAp(ctx context.Context, id *url.URL) error
	CollectionContains(ctx context.Context, collection, id *url.URL) (bool, error)
	GetCollectionPage(ctx context.Context, iri *url.URL, last int64) (ids []*url.URL, err error)
}
