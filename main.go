package main

import (
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"net/url"

	"github.com/alexedwards/scs"
	"github.com/go-chi/chi/v5"
	"github.com/sidereusnuntius/wiki/internal/config"
	db "github.com/sidereusnuntius/wiki/internal/db/impl"
	"github.com/sidereusnuntius/wiki/internal/domain"
	service "github.com/sidereusnuntius/wiki/internal/service/impl"
	"github.com/sidereusnuntius/wiki/internal/state"
	"github.com/sidereusnuntius/wiki/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

// This is a basic, hard wired configuration that only exists for testing. It will change!
func main() {
	u, _ := url.Parse("http://localhost:8080/")

	connString := "file:test.db"
	d, err := sql.Open("sqlite3", connString)
	if err != nil {
		log.Fatal(err)
	}

	gob.Register(domain.Account{})
	manager := scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")

	config := config.Configuration{
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

	service := service.New(state)
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
