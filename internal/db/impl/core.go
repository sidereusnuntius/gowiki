package core

import (
	"database/sql"
	"log"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/db/impl/queries"
)

type dbImpl struct {
	Config  config.Configuration
	db      *sql.DB
	queries *queries.Queries
	DMP     *diffmatchpatch.DiffMatchPatch
}

func New(config config.Configuration, d *sql.DB) db.DB {
	return &dbImpl{
		Config:  config,
		db:      d,
		queries: queries.New(d),
		DMP:     diffmatchpatch.New(),
	}
}

// HandleError takes a database error and returns a higher level error that hides the implementation details
// and can be more easily handled by the calling functions without doing type assertions, checking error codes and
// comparing to sentinel errors.
func (d *dbImpl) HandleError(err error) error {
	switch err {
	case sql.ErrNoRows:
		return db.ErrNotFound
	default:
		if err != nil {
			log.Print(err)
		}
		return err
	}
}
