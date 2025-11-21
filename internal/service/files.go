package service

import (
	"context"
	"net/url"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type FileService interface {
	CreateFile(ctx context.Context, content []byte, metadata domain.FileMetadata) (uri *url.URL, id int64, err error)
	GetFile(ctx context.Context, digest string) (content []byte, metadata domain.File, err error)
}