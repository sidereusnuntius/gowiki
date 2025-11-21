package core

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/service"
	"github.com/sidereusnuntius/gowiki/internal/state"
	"github.com/sidereusnuntius/gowiki/internal/storage/filestore"
)

const (
	RsaKeySize = 2048
	BcryptCost = 10
)

type AppService struct {
	fileServiceImpl
	Config config.Configuration
	DB     db.DB
	DMP    *diffmatchpatch.DiffMatchPatch
}

func New(state *state.State) (service.Service, error) {
	dmp := diffmatchpatch.New()
	store, err := filestore.New(state.Config.FsRoot)
	return &AppService{
		fileServiceImpl: fileServiceImpl{state, store, state.DB},
		Config: state.Config,
		DB:     state.DB,
		DMP:    dmp,
	}, err
}
