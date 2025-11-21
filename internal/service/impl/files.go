package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/state"
	"github.com/sidereusnuntius/gowiki/internal/storage"
)

type fileServiceImpl struct {
	state *state.State
	storage storage.Storage
	DB db.Files
}

func (s *fileServiceImpl) CreateFile(ctx context.Context, content io.Reader, metadata domain.FileMetadata) (id int64, err error) {
	hasher := sha256.New()
	r, w := io.Pipe()

	tr := io.TeeReader(content, hasher)

	go func() {
		_, err = io.Copy(w, tr)
		w.Close()
	}()
	
	
	hash := hex.EncodeToString(hasher.Sum(nil))
	//TODO: verify if hash already exists.
	err = s.storage.Create(r, hash)
	if err != nil {
		return
	}

	uri := s.state.Config.Url.JoinPath("f", hash)
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