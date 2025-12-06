package db

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Users interface {
	GetProfile(ctx context.Context, name string, host sql.NullString) (p domain.Profile, err error)
	GetUserIdByIRI(ctx context.Context, IRI *url.URL) (int64, error)
}
