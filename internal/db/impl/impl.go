package impl

import (
	"database/sql"

	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/db/impl/queries"
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

// WithTx runs a function inside a transaction. It already performs translation of the database errors to the
// higher level sentinel errors defined in the db package.
func (d *dbImpl) WithTx(f func(tx *queries.Queries) error) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		log.Error().Err(err).Msg("unable to begin transaction")
		return db.ErrInternal
	}

	defer func() {
		switch r := recover(); {
		case r != nil:
			fallthrough
		case err != nil:
			_ = tx.Rollback()
		default:
			err = tx.Commit()
		}

		err = d.HandleError(err)
	}()

	err = f(d.queries.WithTx(tx))
	return
}
