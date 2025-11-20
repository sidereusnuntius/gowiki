package state

import (
	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db"
)

// State is a struct holding global objects relevant to many parts of the application. Probably needs refactoring
type State struct {
	DB     db.DB
	Config config.Configuration
}
