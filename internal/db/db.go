package db

import (
	"database/sql"
	"errors"
	"log"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db/queries"
	"github.com/sidereusnuntius/wiki/internal/state"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("")
)

type DB struct {
	Config  config.Configuration
	db      *sql.DB
	queries *queries.Queries
	DMP     *diffmatchpatch.DiffMatchPatch
}

func New(state state.State) DB {
	return DB{
		Config:  state.Config,
		db:      state.DB,
		queries: queries.New(state.DB),
		DMP:     diffmatchpatch.New(),
	}
}

// HandleError takes a database error and returns a higher level error that hides the implementation details
// and can be more easily handled by the calling functions without doing type assertions, checking error codes and
// comparing to sentinel errors.
func (d *DB) HandleError(err error) error {
	switch err {
	case sql.ErrNoRows:
		return ErrNotFound
	default:
		log.Print(err)
		return err
	}
}
