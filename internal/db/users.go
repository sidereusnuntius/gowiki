package db

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Users interface {
	GetUser(ctx context.Context, username, hostname string) (user domain.UserCore, err error)
	GetProfile(ctx context.Context, username, hostname string) (p domain.Profile, err error)
}
