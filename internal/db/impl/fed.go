package impl

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
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

func (d *dbImpl) GetInstanceIdOrCreate(ctx context.Context, hostname string) (id int64, err error) {
	id, err = d.queries.GetInstanceId(ctx, hostname)

	if err == sql.ErrNoRows {
		id, err = d.queries.InsertInstance(ctx, queries.InsertInstanceParams{
			Hostname: hostname,
			PublicKey: sql.NullString{},
			Inbox: sql.NullString{},
		})
	}
	
	if err != nil {
		err = d.HandleError(err)
	}

	return
}

func (d *dbImpl) GetApObject(ctx context.Context, iri *url.URL) (domain.FedObj, error) {
	obj, err := d.queries.GetApObject(ctx, iri.String())
	if err != nil {
		err = d.HandleError(err)
	}

	return domain.FedObj{
		Iri: iri,
		RawJSON: obj.RawJson.String,
		ApType: obj.Type,
		Local: !obj.LastFetched.Valid,
		LocalTable: obj.LocalTable.String,
		LocalId: obj.LastFetched.Int64,
	}, err
}

func (d *dbImpl) GetUserByID(ctx context.Context, id int64) (user domain.UserFed, err error) {
	u, err := d.queries.GetUserFullByID(ctx, id)

	if err != nil {
		err = d.HandleError(err)
		return
	}

	apId, err := url.Parse(u.ApID)
	if err != nil {
		err = db.ErrInternal
		return
	}

	inbox, err := url.Parse(u.Inbox)
	if err != nil {
		err = db.ErrInternal
		return
	}

	outbox, err := url.Parse(u.Outbox)
	if err != nil {
		err = db.ErrInternal
		return
	}

	followers, err := url.Parse(u.Followers)
	if err != nil {
		err = db.ErrInternal
		return
	}
	
	return domain.UserFed{
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
	}, err
}