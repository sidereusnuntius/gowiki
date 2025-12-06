package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"code.superseriousbusiness.org/activity/pub"
	"code.superseriousbusiness.org/activity/streams"
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
	"github.com/sidereusnuntius/gowiki/internal/gateway"
	"github.com/sidereusnuntius/gowiki/internal/initialization"
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

	queue := gateway.New(context.Background(), dd, client, &config, q)

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

	ap := federation.ApService{
		DB: dd,
	}
	actor := pub.NewFederatingActor(&ap, &ap, &fd, Clock{})

	handler := web.New(&config, service, manager)
	router := chi.NewRouter()
	handler.Mount(router)
	wellknown.Mount(&state, router)

	apMux := chi.NewMux()

	apMux.Post("/inbox", func(w http.ResponseWriter, r *http.Request) {
		success, err := actor.PostInbox(r.Context(), w, r)
		e := zero.Debug().Bool("success", success)
		if err != nil {
			e.Err(err)
		}
		e.Msg("attempted to post to inbox.")
	})

	apMux.Get("/outbox", func(w http.ResponseWriter, r *http.Request) {
		r.URL = config.Url.ResolveReference(r.URL)

		if r.URL.Query().Get("last") == "" {
			if err = ap.GetCollection(r.Context(), w, r); err != nil {
				http.Error(w, "", http.StatusInternalServerError)
				zero.Error().Err(err).Msg("trying to get outbox")
			}
			return
		}
		success, err := actor.GetOutbox(r.Context(), w, r)
		e := zero.Debug().Bool("success", success)
		if err != nil {
			e.Err(err)
		}
		e.Msg("attempted to get outbox.")
	})

	apMux.Post("/outbox", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		decoder := json.NewDecoder(r.Body)
		var raw map[string]any
		err := decoder.Decode(&raw)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		activity, err := streams.ToType(ctx, raw)
		if err != nil {
			http.Error(w, "invalid activity", http.StatusBadRequest)
			return
		}

		err = queue.ProcessOutbox(ctx, activity)
		if err != nil {
			zero.Error().Err(err).Msg("failed to post to outbox")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	})

	apMux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		isAp, err := fh(r.Context(), w, r)
		if err != nil {
			zero.Error().Err(err).Send()
			http.Error(w, "", http.StatusInternalServerError)
		}

		if !isAp {
			zero.Error().Msg("not AP request")
		}
	})

	contentRouter := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		content := r.Header.Get("Content-Type")
		// Change this. Please.
		if strings.Contains(accept, "activity") || strings.Contains(content, "activity") || strings.Contains(accept, "ld") || strings.Contains(content, "ld") {
			zero.Log().Str("url", r.URL.String()).Send()
			apMux.ServeHTTP(w, r)
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
