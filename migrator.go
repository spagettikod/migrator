package migrator

import (
	"database/sql"
	"errors"
)

type Migrator interface {
	// Init will set up the Migrator for the current database.
	Init(db *sql.DB) error
	// Initialized will check if the Migrator is setup in this database.
	Initialized(db *sql.DB) (bool, error)
	// Version returns the current version from the database.
	Version() (int, error)
	// SetVersion updates the current version in the database.
	SetVersion(version int) error
	// Migrate will run the forward migrations in the array.
	Migrate(migrations []string) error
}

var ErrMigratorNotInitialized = errors.New("migrator: not initialized, did you call Init?")
