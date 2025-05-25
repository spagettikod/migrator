package migrator

import (
	"database/sql"
	"errors"
)

type SqliteMigrator struct {
	base
}

// NewSqliteMigrator returns a SqliteMigrator ready to run migrations. It will initialize and
// validate the database is ready for migrations. It also validates the given migration target is
// valid.
func NewSqliteMigrator(db *sql.DB) (SqliteMigrator, error) {
	migrations, err := load()
	if err != nil {
		return SqliteMigrator{}, err
	}
	base, err := newBase(db, migrations)
	if err != nil {
		return SqliteMigrator{}, err
	}
	sm := SqliteMigrator{base: base}
	if err := sm.init(); err != nil {
		return sm, err
	}
	return sm, nil
}

// Version returns the current version from the database.
func (sm SqliteMigrator) Version() (int, error) {
	initialized, err := sm.initialized()
	if err != nil {
		return -1, err
	}
	if !initialized {
		return 0, ErrMigratorNotInitialized
	}

	row := sm.db.QueryRow("SELECT version FROM _migrator_")
	version := 0
	if err := row.Scan(&version); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}
	return version, nil
}

// Migrate will migrate the database to version given by the environment variable MIGRATOR_TARGET_VERSION.
// If it fails it will return an error. On successful migration it will return an array with Migration
// that were run.
func (sm SqliteMigrator) Migrate() ([]Migration, error) {
	return sm.migrate(sm)
}

func (sm SqliteMigrator) init() error {
	initialized, err := sm.initialized()
	if err != nil {
		return err
	}
	if !initialized {
		_, err = sm.db.Exec("CREATE TABLE _migrator_ (version INTEGER NOT NULL) STRICT")
		if err != nil {
			return err
		}
		_, err = sm.db.Exec("INSERT INTO _migrator_ (version) VALUES (0)")
	}
	return err
}

func (sm SqliteMigrator) initialized() (bool, error) {
	row := sm.db.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = '_migrator_'")
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

func (sm SqliteMigrator) setVersion(version int) error {
	stmt := "UPDATE _migrator_ SET version = ?1"
	_, err := sm.db.Exec(stmt, version)
	return err
}
