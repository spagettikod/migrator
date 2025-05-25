// Migrator provides an easy way to handle SQL database migrations from Go applications.
//
// Example:
//
//	package main
//
//	import (
//		"database/sql"
//		"log"
//
//		_ "github.com/mattn/go-sqlite3"
//		"github.com/spagettikod/migrator"
//	)
//
//	func main() {
//		// Setup your database connection for your application
//		db, err := sql.Open("sqlite3", ":memory:")
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Create a migrator for your database
//		// (it will initialize it self if run for the first time)
//		migrator, err := migrator.NewSqliteMigrator(db)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		// Run the migrations in up/down to the given target version
//		_, err = migrator.Migrate()
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
package migrator

import (
	"database/sql"
	"errors"
	"os"
	"slices"
	"strconv"
)

const (
	envVarFile              = "MIGRATOR_FILE"
	envVarTarget            = "MIGRATOR_TARGET_VERSION"
	targetStart             = 0
	invalidTarget           = -2
	directionUp   direction = 1
	directionDown direction = 2
	directionNone direction = 0
)

var (
	ErrInvalidTargetVersion    = errors.New("migrator: target version is not valid, make sure MIGRATOR_TARGET_VERSION is correct")
	ErrTargetOutOfBounds       = errors.New("migrator: MIGRATOR_TARGET_VERSION does not match number of migrations")
	ErrMigratorNotInitialized  = errors.New("migrator: not initialized, did you call Init?")
	ErrMigrationFileEnvMissing = errors.New("migrator: environment variable MIGRATOR_FILE empty, can not load migrations")
)

type direction int

type Migrator interface {
	// Version returns the current version from the database.
	Version() (int, error)
	// Migrate will run the forward migrations in the array.
	Migrate() ([]Migration, error)
	// init will set up the Migrator for the current database.
	init() error
	// initialized will check if the Migrator is setup in this database.
	initialized() (bool, error)
	// setVersion updates the current version in the database.
	setVersion(version int) error
}

type base struct {
	db         *sql.DB
	migrations Migrations
	target     int
}

func newBase(db *sql.DB, migrations Migrations) (base, error) {
	migrations.enumerateMigrations()
	b := base{db: db, migrations: migrations}
	target, err := b.parseTarget()
	if err != nil {
		return b, err
	}
	b.target = target
	return b, nil
}

func (b base) parseTarget() (int, error) {
	target := invalidTarget
	tStr, found := os.LookupEnv(envVarTarget)
	if found {
		var err error
		target, err = strconv.Atoi(tStr)
		if err != nil {
			return invalidTarget, ErrInvalidTargetVersion
		}
	}
	if !b.validTarget(target) {
		if target > len(b.migrations.Migrations) {
			return invalidTarget, ErrTargetOutOfBounds
		}
		return invalidTarget, ErrInvalidTargetVersion
	}
	return target, nil
}

func (b base) validTarget(target int) bool {
	if target == invalidTarget {
		return false
	}
	if target < targetStart {
		return false
	}
	return target <= len(b.migrations.Migrations)
}

func migrationDirection(version, target int) direction {
	if version == target {
		return directionNone
	}
	if version > target {
		return directionDown
	}
	return directionUp
}

func (b base) targetMigrations(currVer int) []Migration {
	switch migrationDirection(currVer, b.target) {
	case directionUp:
		return b.migrations.Migrations[currVer:b.target]
	case directionDown:
		revMigations := b.migrations.Migrations[b.target:currVer]
		slices.Reverse(revMigations)
		return revMigations
	default:
		return []Migration{}
	}
}

func (b base) migrate(m Migrator) ([]Migration, error) {
	v, err := m.Version()
	if err != nil {
		return nil, err
	}
	tms := b.targetMigrations(v)
	for _, tm := range tms {
		if _, err := b.db.Exec(tm.stmt(migrationDirection(v, b.target))); err != nil {
			return nil, err
		}
		if migrationDirection(v, b.target) == directionDown {
			m.setVersion(tm.version - 1)
		} else {
			m.setVersion(tm.version)
		}
	}
	return tms, nil
}
