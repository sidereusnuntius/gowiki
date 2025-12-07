package impl

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/gowiki/internal/domain"
)

func (d *dbImpl) IsAdmin(ctx context.Context, accountId int64) (bool, error) {
	isAdmin, err := d.queries.IsAdmin(ctx, accountId)
	if err != nil {
		err = fmt.Errorf("can't verify admin status for id=%d: %w", accountId, d.HandleError(err))
	}
	return isAdmin, err
}

func (d *dbImpl) GetUserURI(ctx context.Context, id int64) (*url.URL, error) {
	uriStr, err := d.queries.GetUserUriById(ctx, id)
	if err != nil {
		return nil, d.HandleError(err)
	}

	uri, err := url.Parse(uriStr)
	return uri, err
}

func (d *dbImpl) UserExists(ctx context.Context, id *url.URL) (exists bool, err error) {
	exists, err = d.queries.UserExists(ctx, id.String())
	if err != nil {
		err = d.HandleError(err)
	}
	return
}

func (d *dbImpl) IsUserTrusted(ctx context.Context, id int64) (bool, error) {
	trusted, err := d.queries.IsUserTrusted(ctx, id)
	return trusted, d.HandleError(err)
}

func (d *dbImpl) GetAuthDataByUsername(ctx context.Context, username string) (domain.Account, error) {
	u, err := d.queries.AuthUserByUsername(ctx, sql.NullString{
		Valid:  true,
		String: username,
	})
	if err != nil {
		return domain.Account{}, err
	}

	apId, err := url.Parse(u.ApID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("%w: unable to parse IRI: %s", db.ErrInternal, u.ApID)
	}
	return domain.Account{
		UserID:    u.UserID,
		AccountID: u.AccountID,
		ApId:      apId,
		Username:  u.Username.String,
		Password:  u.Password,
		Admin:     u.Admin,
	}, nil
}

func (d *dbImpl) GetAuthDataByEmail(ctx context.Context, email string) (domain.Account, error) {
	u, err := d.queries.AuthUserByEmail(ctx, email)
	if err != nil {
		// Treat error to hide implementation details.
		return domain.Account{}, err
	}

	apId, err := url.Parse(u.ApID)
	if err != nil {
		return domain.Account{}, fmt.Errorf("%w: unable to parse IRI: %s", db.ErrInternal, u.ApID)
	}
	return domain.Account{
		UserID:    u.UserID,
		AccountID: u.AccountID,
		ApId:      apId,
		Username:  u.Username.String,
		Password:  u.Password,
		Admin:     u.Admin,
	}, nil
}

func (d *dbImpl) InsertUser(ctx context.Context, user domain.UserFedInternal, account domain.Account, reason string, invitation string) (err error) {
	// TODO: validate and process the invitation, if needed.
	return d.WithTx(func(tx *queries.Queries) error {
		id, err := tx.CreateLocalUser(ctx, queries.CreateLocalUserParams{
			Trusted: user.Trusted,
			ApID:    user.ApId.String(),
			Username: sql.NullString{
				Valid:  user.Username != "",
				String: user.Username,
			},
			Host: sql.NullString{
				Valid:  user.Host != "",
				String: user.Host,
			},
			Name: sql.NullString{
				Valid:  user.Name != "",
				String: user.Name,
			},
			Inbox:      user.Inbox.String(),
			Outbox:     user.Outbox.String(),
			Followers:  user.Followers.String(),
			PublicKey:  user.PublicKey,
			PrivateKey: user.PrivateKey,
			Summary: sql.NullString{
				Valid:  user.Summary != "",
				String: user.Summary,
			},
		})
		if err != nil {
			return err
		}

		err = tx.CreateAccount(ctx, queries.CreateAccountParams{
			Password: account.Password,
			Admin:    account.Admin,
			Email:    account.Email,
			UserID:   id,
		})
		if err != nil {
			return err
		}

		err = tx.InsertApObject(ctx, queries.InsertApObjectParams{
			ApID: user.ApId.String(),
			LocalTable: sql.NullString{
				Valid:  true,
				String: "users",
			},
			LocalID: sql.NullInt64{
				Valid: true,
				Int64: id,
			},
			Type: "Person",
		})
		if err != nil {
			return err
		}

		if err = insertCollection(ctx, tx, user.Inbox); err != nil {
			return err
		}
		if err = insertCollection(ctx, tx, user.Outbox); err != nil {
			return err
		}
		return insertCollection(ctx, tx, user.Followers)
	})
}

func insertCollection(ctx context.Context, tx *queries.Queries, iri *url.URL) error {
	return tx.InsertApObject(ctx, queries.InsertApObjectParams{
		ApID: iri.String(),
		Type: "OrderedCollection",
	})
}
