// The init package contains functions that setup required dependencies such as the SQLite database.
package initialization

import (
	"database/sql"
	
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/rs/zerolog/log"
)

// SetupDB creates the database, if it does not yet exist, and applies all remaining migrations, then closes the
// connection.
func SetupDB(db *sql.DB, dbname string) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create sqlite3 migration driver")
		db.Close()
		return err
	}
	
	mig, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
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
	}
	return err
}

func OpenDB(connString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		log.Fatal().Err(err).Str("connection string", connString).Msg("failed to open database")
	}
	return db, err
}