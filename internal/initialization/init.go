// The init package contains functions that setup required dependencies such as the SQLite database.
package initialization

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mikestefanello/backlite"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/utils"
)

func InitQueue(cfg *config.Configuration) (client *backlite.Client, err error) {
	db, err := sql.Open("sqlite3", cfg.QueueDbPath+"?_journal=WAL")
	if err != nil {
		return
	}

	client, err = backlite.NewClient(backlite.ClientConfig{
		DB:              db,
		Logger:          slog.Default(),
		NumWorkers:      6,
		ReleaseAfter:    10 * time.Minute,
		CleanupInterval: time.Hour,
	})
	if err != nil {
		return
	}

	err = client.Install()
	return
}

// SetupDB creates the database, if it does not yet exist, and applies all remaining migrations, then closes the
// connection.
func SetupDB(cfg *config.Configuration, db *sql.DB, folder, dbname string) error {
	log.Info().Msg("starting migrations")
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create sqlite3 migration driver")
		db.Close()
		return err
	}

	mig, err := migrate.NewWithDatabaseInstance(
		"file://"+folder,
		dbname,
		driver,
	)

	if err != nil {
		log.Fatal().Err(err).Msg("failed to create Migrate object")
		return err
	}

	err = mig.Up()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
		return err
	}

	return EnsureInstance(db, cfg)
}

func OpenDB(connString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		log.Fatal().Err(err).Str("connection string", connString).Msg("failed to open database")
	}
	return db, err
}

func EnsureInstance(DB *sql.DB, cfg *config.Configuration) error {
	row := DB.QueryRow("SELECT EXISTS(SELECT TRUE FROM instances WHERE url = ?)", cfg.Url.String())
	var exists bool
	err := row.Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}
	log.Info().Msg("inserting server data into the database")
	inbox := cfg.Url.JoinPath("inbox").String()
	outbox := cfg.Url.JoinPath("outbox").String()
	followers := cfg.Url.JoinPath("followers").String()

	pub, priv, err := utils.GenerateKeysPem(2048)
	if err != nil {
		return err
	}

	res, err := DB.Exec(`INSERT INTO instances(
				name,
				hostname,
				url,
				public_key,
				private_key,
				inbox,
				outbox,
				followers
			) VALUES (?,?,?,?,?,?,?,?) RETURNING id`,
		cfg.Name, cfg.Url.Host, cfg.Url.String(), pub, priv, inbox, outbox, followers)
	if err != nil {
		log.Error().Err(err).Msg("insert failed")
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Error().Err(err).Msg("could not get last inserted id")
		return err
	}

	_, err = DB.Exec(`
INSERT INTO ap_object_cache (ap_id, local_table, local_id, type)
VALUES (?, ?, ?, ?)`, cfg.Url.String(), "instances", id, "Group")
	return err
}
