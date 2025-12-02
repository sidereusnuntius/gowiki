package impl

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (d *dbImpl) GetUser(ctx context.Context, username, hostname string) (user domain.UserCore, err error) {
	if hostname == "" {
		var u queries.GetLocalUserDataRow
		u, err = d.queries.GetLocalUserData(ctx, username)
		uri, _ := url.Parse(u.Url.String)

		user = domain.UserCore{
			ID:       u.ID,
			Username: u.Username.String,
			Name:     u.Name.String,
			Domain:   "",
			Summary:  u.Summary.String,
			URL:      uri,
		}
	} else {
		var u queries.GetForeignUserDataRow
		u, err = d.queries.GetForeignUserData(
			ctx,
			queries.GetForeignUserDataParams{LOWER: username, Host: sql.NullString{ Valid: true, String: hostname}},
		)

		uri, _ := url.Parse(u.Url.String)
		user = domain.UserCore{
			ID:       u.ID,
			Username: u.Username.String,
			Name:     u.Name.String,
			Domain:   u.Host.String,
			Summary:  u.Summary.String,
			URL:      uri,
		}
	}

	if err != nil {
		err = d.HandleError(err)
	}

	return
}

func (d *dbImpl) GetProfile(ctx context.Context, username, hostname string) (p domain.Profile, err error) {
	u, err := d.GetUser(ctx, username, hostname)
	if err != nil {
		return
	}

	r, err := d.queries.GetRevisionsByUserId(ctx, u.ID)
	edits := make([]domain.Revision, 0, len(r))
	for _, r := range r {
		edits = append(edits, domain.Revision{
			ID:       r.ID,
			Title:    r.Title,
			Reviewed: r.Reviewed,
			Summary:  r.Summary.String,
			Created:  r.Created,
		})
	}
	p = domain.Profile{
		UserCore: u,
		Edits:    edits,
	}

	if err != nil {
		err = d.HandleError(err)
	}
	return
}
