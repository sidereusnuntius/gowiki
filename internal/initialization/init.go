// The init package contains functions that setup required dependencies such as the SQLite database.
package initialization

import (
	"database/sql"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/sidereusnuntius/gowiki/internal/config"
	"github.com/sidereusnuntius/gowiki/internal/utils"
)

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

	_, err = DB.Exec(`INSERT INTO instances(
				name,
				hostname,
				url,
				public_key,
				private_key,
				inbox,
				outbox,
				followers
			) VALUES (?,?,?,?,?,?,?,?)`,
		cfg.Name, cfg.Url.Host, cfg.Url.String(), pub, priv, inbox, outbox, followers)
	if err != nil {
		log.Error().Err(err).Msg("insert failed")
	}
	return err
}
