package impl

import (
	"context"
	"net/url"
	"testing"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	"github.com/sidereusnuntius/gowiki/internal/initialization"
)

var DB db.DB
var ctx = context.Background()

func TestMain(m *testing.M) {
	d, err := initialization.OpenDB("file:temp?mode=memory")
	if err != nil {
		return
	}

	err = initialization.SetupDB(d, "temp")
	if err != nil {
		return
	}
	hostname, _ := url.Parse("https://test.wiki")
	DB = New(config.Configuration{
	    Domain: "test.wiki",
		Url: hostname,
	}, d)
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