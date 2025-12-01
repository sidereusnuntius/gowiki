package main

import (
	"context"
	"encoding/gob"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/httpsig"
	"github.com/alexedwards/scs"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	zero "github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/client"
	"github.com/sidereusnuntius/gowiki/internal/config"
	db "github.com/sidereusnuntius/gowiki/internal/db/impl"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
	"github.com/sidereusnuntius/gowiki/internal/federation/fedb"
	"github.com/sidereusnuntius/gowiki/internal/initialization"
	"github.com/sidereusnuntius/gowiki/internal/queue"
	service "github.com/sidereusnuntius/gowiki/internal/service/impl"
	"github.com/sidereusnuntius/gowiki/internal/state"
	"github.com/sidereusnuntius/gowiki/internal/web"
	"github.com/sidereusnuntius/gowiki/internal/wellknown"

	_ "github.com/mattn/go-sqlite3"
)

type Clock struct{}

func (c Clock) Now() time.Time {
	return time.Now()
}

// This is a basic, hard wired configuration that only exists for testing. It will change!
func main() {
	zero.Logger = zero.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	config, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	d, err := initialization.OpenDB(config.DbUrl)
	if err != nil {
		log.Fatal(err)
	}
	zero.Info().Msg("database connection established")

	q, err := initialization.InitQueue(&config)
	if err != nil {
		zero.Fatal().Err(err).Msg("unable to connect with backlite database")
		os.Exit(1)
	}

	if os.Getenv("SETUP") != "" {
		err = initialization.SetupDB(&config, d, config.MigrationsFolder, config.DbUrl)
		if err != nil {
			log.Fatal(err)
		}
	}

	gob.Register(domain.Account{})
	manager := scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")

	err = initialization.EnsureInstance(d, &config)
	if err != nil {
		log.Fatal(err)
	}

	dd := db.New(config, d)
	key, err := dd.GetUserPrivateKeyByURI(context.Background(), config.Url)
	if err != nil {
		zero.Fatal().Err(err).Send()
		os.Exit(1)
	}

	fragment, _ := url.Parse("#main-key")
	keyId := config.Url.ResolveReference(fragment)
	client, err := client.New(dd, &http.Client{}, key, []httpsig.Algorithm{httpsig.RSA_SHA256}, keyId)
	if err != nil {
		zero.Fatal().Err(err).Send()
		os.Exit(1)
	}

	queue := queue.New(context.Background(), dd, client, &config, q)

	state := state.State{
		DB:     dd,
		Config: config,
	}

	service, err := service.New(&state, queue)
	if err != nil {
		log.Fatal(err)
	}

	fd := fedb.New(state.DB, queue, config)
	fh := pub.NewActivityStreamsHandler(&fd, Clock{})

	ap := federation.ApService{}
	actor := pub.NewFederatingActor(&ap, &ap, &fd, Clock{})

	handler := web.New(&config, service, manager)
	router := chi.NewRouter()
	handler.Mount(router)
	wellknown.Mount(&state, router)
	router.Post("/inbox", func(w http.ResponseWriter, r *http.Request) {
		success, err := actor.PostInbox(r.Context(), w, r)
		recover()
		e := zero.Debug().Bool("success", success)
		if err != nil {
			e.Err(err)
		}
		e.Msg("attempted to post to inbox.")
	})

	contentRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "application/ld+json") || strings.Contains(accept, "application/activity+json") {
			zero.Log().Str("url", r.URL.String()).Send()
			isAp, err := fh(r.Context(), w, r)
			if err != nil {
				zero.Error().Err(err).Send()
				http.Error(w, "", http.StatusInternalServerError)
			}
			if !isAp {
				http.Error(w, "could not process request", http.StatusInternalServerError)
			}
		} else {
			router.ServeHTTP(w, r)
		}
	})

	if config.Debug {
		// Register logging middleware.
	}

	s := &http.Server{
		Addr:    ":8080",
		Handler: contentRouter,
	}

	zero.Info().Uint16("port", config.Port).Msg("started server")
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
