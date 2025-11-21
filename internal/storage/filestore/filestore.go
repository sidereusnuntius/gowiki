package filestore

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/storage"
)

type FileStore struct {
	Root string
}

func New(root string) (fs storage.Storage, err error) {
	fs = &FileStore{
		Root: root,
	}

	info, err := os.Stat(root)
	if err == nil {
		if !info.IsDir() {
			log.Error().Msg("not a directory")
			err = storage.ErrNotDir
		}
		return
	}

	if errors.Is(err, os.ErrNotExist) {
		err = os.Mkdir(root, os.ModePerm)
	}

	if err != nil {
		log.Error().Err(err).Msg("internal error when setting up storage")
		err = storage.ErrInternal
	}

	return
}

func (s *FileStore) Open(path string) (content []byte, err error) {
	path = filepath.Join(s.Root, path)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = storage.ErrNotExist
		} else {
			err = storage.ErrInternal
			log.Error().Err(err).Msg("failed to open file at path " + path)
		}
		return
	}
	defer f.Close()
	// Perhaps we should add an optional parameter specifying the file size?
	content, err = io.ReadAll(f)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file " + path)
		err = storage.ErrInternal
	}
	return
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
