package db

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type Files interface {
	Save(ctx context.Context, file domain.File) (id int64, err error)
	FileExists(ctx context.Context, hash string) (bool, error)
	GetFile(ctx context.Context, hash string) (domain.File, error)
}
