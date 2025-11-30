package db

import (
	"context"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Account interface {
	// InsertUser hashes the user's password and persists the user in the database; if needed, it will check if
	// the invitation is valid and update it's status so it can't be used again.
	GetUserURI(ctx context.Context, id int64) (*url.URL, error)
	InsertUser(ctx context.Context, user domain.UserFedInternal, account domain.Account, reason string, invitation string) (err error)
	UserExists(ctx context.Context, id *url.URL) (exists bool, err error)
	IsUserTrusted(ctx context.Context, id int64) (bool, error)
	GetAuthDataByUsername(ctx context.Context, username string) (domain.Account, error)
	GetAuthDataByEmail(ctx context.Context, email string) (domain.Account, error)
}
