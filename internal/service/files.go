package service

import (
	"context"
	"io"

	"github.com/sidereusnuntius/gowiki/internal/domain"
)

type FileService interface {
	CreateFile(ctx context.Context, content io.Reader, metadata domain.FileMetadata) (id int64, err error)
}