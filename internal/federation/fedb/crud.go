package fedb

import (
	"context"
	"net/url"

	"code.superseriousbusiness.org/activity/streams/vocab"
)

func (fd *FedDB) Get(ctx context.Context, id *url.URL) (value vocab.Type, err error) {

	return
}

func (fd *FedDB) Create(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (fd *FedDB) Update(ctx context.Context, asType vocab.Type) error {
	return nil
}

func (fd *FedDB) Delete(ctx context.Context, id *url.URL) error {
	return nil
}
