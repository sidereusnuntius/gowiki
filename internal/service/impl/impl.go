package core

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/service"
	"github.com/sidereusnuntius/gowiki/internal/state"
)

const (
	RsaKeySize = 2048
	BcryptCost = 10
)

type AppService struct {
	Config config.Configuration
	DB     db.DB
	DMP    *diffmatchpatch.DiffMatchPatch
}

func New(state state.State) service.Service {
	dmp := diffmatchpatch.New()
	return &AppService{
		Config: state.Config,
		DB:     state.DB,
		DMP:    dmp,
	}
}
