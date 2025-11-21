package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/alexedwards/scs"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	zero "github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/config"
	db "github.com/sidereusnuntius/gowiki/internal/db/impl"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	service "github.com/sidereusnuntius/gowiki/internal/service/impl"
	"github.com/sidereusnuntius/gowiki/internal/state"
	"github.com/sidereusnuntius/gowiki/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

// This is a basic, hard wired configuration that only exists for testing. It will change!
func main() {
	zero.Logger = zero.Output(zerolog.ConsoleWriter{ Out: os.Stderr })
	u, _ := url.Parse("http://localhost:8080/")

	connString := "file:test.db"
	d, err := sql.Open("sqlite3", connString)
	if err != nil {
		log.Fatal(err)
	}

	gob.Register(domain.Account{})
	manager := scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")

	config := config.Configuration{
		FsRoot: "./files",
		StaticDir:          "/static/",
		RsaKeySize:         2048,
		InvitationRequired: false,
		ApprovalRequired:   false,
		Https:              false,
		Debug:              true,
		Domain:             "localhost:8080",
		DbUrl:              connString,
		Url:                u,
	}

	dd := db.New(config, d)

	state := state.State{
		DB:     dd,
		Config: config,
	}

	service, err := service.New(&state)
	if err != nil {
		log.Fatal(err)
	}
	handler := web.New(&config, service, manager)
	r := chi.NewRouter()
	handler.Mount(r)
	if config.Debug {
		// Register logging middleware.
	}

	s := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
