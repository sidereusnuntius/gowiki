package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/service"
	"github.com/sidereusnuntius/gowiki/internal/state"
	"github.com/sidereusnuntius/gowiki/internal/storage"
)

type fileServiceImpl struct {
	state *state.State
	storage storage.Storage
	DB db.Files
}

func (s *fileServiceImpl) CreateFile(ctx context.Context, content []byte, metadata domain.FileMetadata) (uri *url.URL, id int64, err error) {
	hasher := sha256.New()
	n, err := hasher.Write(content)
	if n != int(metadata.SizeBytes) {
		err = service.ErrInvalidInput
		return
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	exists, err := s.DB.FileExists(ctx, hash)
	if err != nil {
		return
	}

	if exists {
		err = service.ErrConflict
		return
	}

	//TODO: verify if hash already exists.
	err = s.storage.Create(bytes.NewReader(content), hash)
	if err != nil {
		return
	}

	uri = s.state.Config.Url.JoinPath("f", hash)
	id, err = s.DB.Save(ctx, domain.File{
		FileMetadata: metadata,
		Digest: hash,
		Path: hash,
		ApId: uri,
		Url: uri,
	})

	if err != nil {
		if err := s.storage.Delete(hash); err != nil {
			log.Error().
				Str("path", hash).
				Str("type", metadata.MimeType).
				Err(err).
				Msg("error when trying to delete file after failed transaction")
		}
	}

	return
}

func (s *fileServiceImpl) GetFile(ctx context.Context, digest string) (content []byte, metadata domain.File, err error) {
	metadata, err = s.DB.GetFile(ctx, digest)
	if err != nil {
		return
	}

	content, err = s.storage.Open(digest)
	return
}