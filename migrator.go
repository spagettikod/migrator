package migrator

import (
	"database/sql"
	"errors"
	"os"
	"slices"
	"strconv"
)

const (
	EnvVarFile              = "MIGRATOR_FILE"
	EnvVarTarget            = "MIGRATOR_TARGET_VERSION"
	TargetStart             = 0
	InvalidTarget           = -2
	DirectionUp   Direction = 1
	DirectionDown Direction = 2
	DirectionNone Direction = 0
)

var ErrInvalidTargetVersion = errors.New("migrator: target version is not valid, make sure MIGRATOR_TARGET_VERSION is correct")
var ErrTargetOutOfBounds = errors.New("migrator: MIGRATOR_TARGET_VERSION does not match number of migrations")
var ErrMigratorNotInitialized = errors.New("migrator: not initialized, did you call Init?")
var ErrMigrationFileEnvMissing = errors.New("migrator: environment variable MIGRATOR_FILE empty, can not load migrations")

type Direction int

type Migrator interface {
	// Init will set up the Migrator for the current database.
	init() error
	// Initialized will check if the Migrator is setup in this database.
	initialized() (bool, error)
	// SetVersion updates the current version in the database.
	setVersion(version int) error
	// Version returns the current version from the database.
	Version() (int, error)
	// Migrate will run the forward migrations in the array.
	Migrate() ([]Migration, error)
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
	target := InvalidTarget
	tStr, found := os.LookupEnv(EnvVarTarget)
	if found {
		var err error
		target, err = strconv.Atoi(tStr)
		if err != nil {
			return InvalidTarget, ErrInvalidTargetVersion
		}
	}
	if !b.validTarget(target) {
		if target > len(b.migrations.Migrations) {
			return InvalidTarget, ErrTargetOutOfBounds
		}
		return InvalidTarget, ErrInvalidTargetVersion
	}
	return target, nil
}

func (b base) validTarget(target int) bool {
	if target == InvalidTarget {
		return false
	}
	if target < TargetStart {
		return false
	}
	return target <= len(b.migrations.Migrations)
}

func direction(version, target int) Direction {
	if version == target {
		return DirectionNone
	}
	if version > target {
		return DirectionDown
	}
	return DirectionUp
}

func (b base) targetMigrations(currVer int) []Migration {
	switch direction(currVer, b.target) {
	case DirectionUp:
		return b.migrations.Migrations[currVer:b.target]
	case DirectionDown:
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
		if _, err := b.db.Exec(tm.Stmt(direction(v, b.target))); err != nil {
			return nil, err
		}
		if direction(v, b.target) == DirectionDown {
			m.setVersion(tm.version - 1)
		} else {
			m.setVersion(tm.version)
		}
	}
	return tms, nil
}
