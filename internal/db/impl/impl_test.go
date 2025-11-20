package impl

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/db"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var DB db.DB
var ctx = context.Background()

func TestMain(m *testing.M) {
	d, err := sql.Open("sqlite3", "file:temp?mode=memory")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open connection: %s", err)
		return
	}

	driver, err := sqlite3.WithInstance(d, &sqlite3.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create driver: %s", err)
		return
	}

	mig, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
		"temp",
		driver,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create database object: %s", err)
		return
	}

	err = mig.Up()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %s", err)
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