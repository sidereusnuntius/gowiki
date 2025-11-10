package state

import (
	"database/sql"

	"github.com/sidereusnuntius/wiki/internal/config"
)

type State struct {
	DB *sql.DB
	Config config.Configuration
}