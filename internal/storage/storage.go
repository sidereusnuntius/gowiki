package storage

import (
	"errors"
	"io"
)

var (
	ErrNotDir = errors.New("given root is not a directory")
	ErrInternal = errors.New("internal error")
	ErrCreate = errors.New("failed to create file")
	ErrAlreadyExists = errors.New("filename already exists")
	ErrNotExist = errors.New("file does not exist")
)

type Storage interface {
	Open(path string) ([]byte, error)
	Create(content io.Reader, path string) error
	Delete(path string) error	
}