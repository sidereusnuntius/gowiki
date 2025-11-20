package core

import (
	"context"
	"net/url"
	"time"

	"github.com/sidereusnuntius/wiki/internal/domain"
)

func (d *dbImpl) ActorIdByOutbox(ctx context.Context, iri *url.URL) (*url.URL, error) {
	id, err := d.queries.UserIdByOutbox(ctx, iri.String())

	if err != nil {
		return nil, d.HandleError(err)
	}
	iri, err = url.Parse(id)

	return iri, d.HandleError(err)
}

func (d *dbImpl) ActorIdByInbox(ctx context.Context, iri *url.URL) (*url.URL, error) {
	id, err := d.queries.UserIdByInbox(ctx, iri.String())

	if err != nil {
		return nil, d.HandleError(err)
	}
	iri, err = url.Parse(id)

	return iri, d.HandleError(err)
}

func (d *dbImpl) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (*url.URL, error) {
	id, err := d.queries.OutboxForInbox(ctx, inboxIRI.String())
	if err != nil {
		return nil, d.HandleError(err)
	}
	outboxIRI, err := url.Parse(id)
	return outboxIRI, d.HandleError(err)
}

func (d *dbImpl) GetUserFed(ctx context.Context, id *url.URL) (user domain.UserFed, err error) {
	defer func() {
		if err != nil {
			err = d.HandleError(err)
		}
	}()

	u, err := d.queries.GetUserFull(ctx, id.String())
	if err != nil {
		return
	}

	apId, err := url.Parse(u.ApID)
	if err != nil {
		return
	}

	inbox, err := url.Parse(u.Inbox)
	if err != nil {
		return
	}

	outbox, err := url.Parse(u.Outbox)
	if err != nil {
		return
	}

	followers, err := url.Parse(u.Followers)
	if err != nil {
		return
	}

	user = domain.UserFed{
		UserCore: domain.UserCore{
			Username: u.Username,
			Name:     u.Name,
			Summary:  u.Summary.String,
			//URL: , TODO
		},
		ApId:        apId,
		Inbox:       inbox,
		Outbox:      outbox,
		Followers:   followers,
		PublicKey:   u.PublicKey,
		Created:     time.Unix(u.Created, 0),
		LastUpdated: time.Unix(u.LastUpdated, 0),
	}
	return
}
