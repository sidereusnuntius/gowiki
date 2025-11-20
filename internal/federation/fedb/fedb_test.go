package fedb

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

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

	m.Run()

	// r := d.QueryRow("SELECT COUNT(id) FROM users")
	// var count int64
	// if err = r.Scan(&count); err != nil {
	// 	fmt.Fprintf(os.Stderr, "unexpected: %s", err)
	// } else {
	// 	fmt.Fprintf(os.Stderr, "%d\n", count)
	// }

	d.Close()
	err, err2 := mig.Close()
	if err != nil || err2 != nil {
		fmt.Fprintf(os.Stderr, "Source: %s\nDatabase: %s\n", err, err2)
		return
	}
}
