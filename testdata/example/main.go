package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spagettikod/migrator"
)

func main() {
	// Setup your database connection for your application
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}

	// Create a migrator for your database
	// (it will initialize it self if run for the first time)
	migrator, err := migrator.NewSqliteMigrator(db)
	if err != nil {
		log.Fatal(err)
	}

	// Run the migrations in up/down to the given target version
	_, err = migrator.Migrate()
	if err != nil {
		log.Fatal(err)
	}
}
