package state

import (
	"net/http"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
)

// State is a struct holding global objects relevant to many parts of the application. Probably needs refactoring
type State struct {
	Client *http.Client
	DB     db.DB
	Config config.Configuration
}
