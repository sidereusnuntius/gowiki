package filestore

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/wiki/internal/storage"
)

type FileStore struct {
	Root string
}

func (s *FileStore) Delete(path string) error {
	path = filepath.Join(s.Root, path)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return storage.ErrNotExist
		}
		log.Error().Err(err).Msg("file deletion error")
		return storage.ErrInternal
	}
	
	return nil
}

func (s *FileStore) Create(content io.Reader, path string) error {
	path = filepath.Join(s.Root, path)
	_, err := os.Stat(path)
	if err == nil {
		return storage.ErrAlreadyExists	
	}
	if !os.IsNotExist(err) {
		log.Error().Err(err).Msg("unknown filesystem error")
		return storage.ErrInternal
	}

	file, err := os.Create(path)
	if err != nil {
		log.Error().Err(err).Msg("failed to create file with path " + path)
		return storage.ErrCreate
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		log.Error().Err(err).Msg("failed to copy from reader")
		return storage.ErrInternal
	}

	return nil
}
