package core

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/db/impl/queries"
	"github.com/sidereusnuntius/wiki/internal/domain"
)

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
	u, err := d.queries.AuthUserByUsername(ctx, username)
	if err != nil {
		return domain.Account{}, err
	}

	return domain.Account{
		UserID:    u.UserID,
		AccountID: u.AccountID,
		Username:  u.Username,
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

	return domain.Account{
		UserID:    u.UserID,
		AccountID: u.AccountID,
		Username:  u.Username,
		Password:  u.Password,
		Admin:     u.Admin,
	}, nil
}

func (d *dbImpl) InsertUser(ctx context.Context, user domain.UserFedInternal, account domain.Account, reason string, invitation string) (err error) {
	// TODO: validate and process the invitation, if needed.
	tx, err := d.db.Begin()
	if err != nil {
		//Log error
		err = db.ErrInternal
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	t := d.queries.WithTx(tx)
	id, err := t.CreateLocalUser(ctx, queries.CreateLocalUserParams{
		Trusted:    user.Trusted,
		ApID:       user.ApId.String(),
		Username:   user.Username,
		Name:       user.Name,
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
		return
	}

	account.UserID = id

	err = t.CreateAccount(ctx, queries.CreateAccountParams{
		Password: account.Password,
		Admin:    account.Admin,
		Email:    account.Email,
		UserID:   account.UserID,
	})

	if d.Config.ApprovalRequired {
		// TODO: insert an approval request.
	}

	return
}
