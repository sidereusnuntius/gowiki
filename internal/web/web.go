package web

import (
	"github.com/alexedwards/scs"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/service"
)

const (
	LoginRoute   = "/login"
	SignUpRoute  = "/signup"
	ArticlesPath = "/a"
)

type Handler struct {
	Config         *config.Configuration
	service        service.Service
	SessionManager *scs.Manager
}

func New(config *config.Configuration, service service.Service, manager *scs.Manager) Handler {
	return Handler{
		Config:         config,
		service:        service,
		SessionManager: manager,
	}
}
