package migrator

import (
	"database/sql"
	"errors"
	"fmt"
)

type PostgresMigrator struct {
	base
	schema string
}

// NewPostgresMigrator returns a PostgresMigrator ready to run migrations. It will initialize and
// validate the database is ready for migrations. It also validates the given migration target is
// valid.
func NewPostgresMigrator(db *sql.DB, schema string) (PostgresMigrator, error) {
	migrations, err := load()
	if err != nil {
		return PostgresMigrator{}, err
	}
	base, err := newBase(db, migrations)
	if err != nil {
		return PostgresMigrator{}, err
	}

	if schema == "" {
		schema = "public"
	}
	sm := PostgresMigrator{base: base, schema: schema}
	if err := sm.init(); err != nil {
		return sm, err
	}
	return sm, nil
}

// Version returns the current version from the database.
func (pm PostgresMigrator) Version() (int, error) {
	initialized, err := pm.initialized()
	if err != nil {
		return -1, err
	}
	if !initialized {
		return 0, ErrMigratorNotInitialized
	}

	row := pm.db.QueryRow(fmt.Sprintf("SELECT version FROM %s._migrator_", pm.schema))
	version := 0
	if err := row.Scan(&version); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}
	return version, nil
}

// Migrate will migrate the database to version given by the environment variable MIGRATOR_TARGET_VERSION.
// If it fails it will return an error. On successful migration it will return an array with Migration
// that were run.
func (pm PostgresMigrator) Migrate() ([]Migration, error) {
	return pm.migrate(pm)
}

func (pm PostgresMigrator) init() error {
	initialized, err := pm.initialized()
	if err != nil {
		return err
	}
	if !initialized {
		_, err = pm.db.Exec(fmt.Sprintf("CREATE TABLE %s._migrator_ (version INTEGER NOT NULL)", pm.schema))
		if err != nil {
			return err
		}
		_, err = pm.db.Exec(fmt.Sprintf("INSERT INTO %s._migrator_ (version) VALUES (0)", pm.schema))
		return err
	}
	return nil
}

func (pm PostgresMigrator) initialized() (bool, error) {
	row := pm.db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_schema = $1 AND table_name = '_migrator_'", pm.schema)
	name := ""
	err := row.Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (pm PostgresMigrator) setVersion(version int) error {
	stmt := fmt.Sprintf("UPDATE %s._migrator_ SET version = $1", pm.schema)
	_, err := pm.db.Exec(stmt, version)
	return err
}
