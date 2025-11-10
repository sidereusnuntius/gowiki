package db

import (
	"database/sql"
	"errors"

	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db/queries"
	"github.com/sidereusnuntius/wiki/internal/state"
)

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("")
)

type DB struct {
	Config config.Configuration
	db *sql.DB
	queries *queries.Queries
}

func New(state state.State) DB {
	return DB {
		Config: state.Config,
		db: state.DB,
		queries: queries.New(state.DB),
	}
}