package db

import (
	"context"
	"net/url"

	"github.com/sidereusnuntius/wiki/internal/domain"
)

type Fed interface {
	ActorIdByInbox(ctx context.Context, iri *url.URL) (*url.URL, error)
	ActorIdByOutbox(ctx context.Context, iri *url.URL) (*url.URL, error)
	OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (*url.URL, error)
	GetUserFed(ctx context.Context, id *url.URL) (user domain.UserFed, err error)
	GetInstanceIdOrCreate(ctx context.Context, hostname string) (id int64, err error)
}
