package service

import (
	"context"
	"errors"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

var (
	ErrConflict     = errors.New("conflict")
	ErrInvalidInput = errors.New("invalid")
)

// Remove the use of sqlc generated and db-defined structs.
type Service interface {
	FileService
	// AuthenticateUser takes the user's identifier, which may be their username of email address, and password
	// and verifies if these credentials are correct. If authentication fails, authenticated is false and
	// err is nil; a non nil error indicates that an internal, unexpected error has occured.
	AuthenticateUser(ctx context.Context, user, password string) (u domain.Account, authenticated bool, err error)
	// CreateUser inserts a new, local user, also creating their corresponding account, for which the email and password
	// are needed.
	CreateUser(ctx context.Context, username, password, email, reason string, admin bool, invitation string) error
	// AlterArticle creates the article if it does not exists; otherwise it will modify the article,
	// recording the edit in the article's history.
	AlterArticle(ctx context.Context, title, summary, content string, userId int64) (*url.URL, error)
	GetLocalArticle(ctx context.Context, title string) (article domain.ArticleCore, err error)
	CreateArticle(ctx context.Context, title, summary, content string, userId int64) (*url.URL, error)
	GetUserProfile(ctx context.Context, username, domain string) (p domain.Profile, err error)
	GetRevisionList(ctx context.Context, title string) ([]domain.Revision, error)
}
