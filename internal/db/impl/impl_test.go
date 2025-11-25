package impl

import (
	"context"
	"net/url"
	"testing"

	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/initialization"
)

var DB db.DB
var ctx = context.Background()

func TestMain(m *testing.M) {
	hostname, _ := url.Parse("https://test.wiki")
	cfg := config.Configuration{
		Domain: "test.wiki",
		Url:    hostname,
	}
	d, err := initialization.OpenDB("file:temp?mode=memory")
	if err != nil {
		return
	}

	err = initialization.SetupDB(&cfg, d, "../../../migrations", "temp")
	if err != nil {
		return
	}
	DB = New(cfg, d)
	m.Run()
}

func TestGetInstanceIdOrCreate(t *testing.T) {
	id1, err := DB.GetInstanceIdOrCreate(ctx, "comp.wiki")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	id2, err := DB.GetInstanceIdOrCreate(ctx, "comp.wiki")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if id1 != id2 {
		t.Errorf("expected second query to return id %d, but it returned %d", id1, id2)
	}
}
