package db

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("")
)

type DB interface {
	Account
	Article
	Fed
	Users
	Files
}
