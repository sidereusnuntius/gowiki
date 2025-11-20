package core

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/wiki/internal/config"
	"github.com/sidereusnuntius/wiki/internal/db"
	"github.com/sidereusnuntius/wiki/internal/service"
	"github.com/sidereusnuntius/wiki/internal/state"
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
