package db

import (
	"context"

	"github.com/sidereusnuntius/wiki/internal/db/queries"
)

type Account interface {
	// InsertUser hashes the user's password and persists the user in the database; if needed, it will check if
	// the invitation is valid and update it's status so it can't be used again.
	InsertUser(ctx context.Context, user queries.CreateLocalUserParams, account queries.CreateAccountParams, reason string, invitation int) (err error)
}

type UserData struct {
	UserID    int64
	AccountID int64
	Username  string
	Password  string
	Admin     bool
}

func (d *DB) IsUserTrusted(ctx context.Context, id int64) (bool, error) {
	trusted, err := d.queries.IsUserTrusted(ctx, id)
	return trusted, d.HandleError(err)
}

func (d *DB) GetAuthDataByUsername(ctx context.Context, username string) (UserData, error) {
	u, err := d.queries.AuthUserByUsername(ctx, username)
	if err != nil {
		return UserData{}, err
	}

	return UserData(u), nil
}

func (d *DB) GetAuthDataByEmail(ctx context.Context, email string) (UserData, error) {
	u, err := d.queries.AuthUserByEmail(ctx, email)
	if err != nil {
		// Treat error to hide implementation details.
		return UserData{}, err
	}

	return UserData(u), nil
}

func (d *DB) InsertUser(ctx context.Context, user queries.CreateLocalUserParams, account queries.CreateAccountParams, reason string, invitation string) (err error) {
	// TODO: validate and process the invitation, if needed.
	tx, err := d.db.Begin()
	if err != nil {
		//Log error
		err = ErrInternal
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
	id, err := t.CreateLocalUser(ctx, user)
	if err != nil {
		return
	}

	account.UserID = id

	err = t.CreateAccount(ctx, account)

	if d.Config.ApprovalRequired {
		// TODO: insert an approval request.
	}

	return
}
