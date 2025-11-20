package federation

import (
	"context"
	"net/http"
)

type FedProto struct {
}

func (f *FedProto) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return
}
