package migrator

import (
	"database/sql"
	"errors"
)

type SqliteMigrator struct {
	base
}

func NewSqliteMigrator(db *sql.DB) (Migrator, error) {
	migrations, err := Load()
	if err != nil {
		return nil, err
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

// Init will set up the Migrator for the current database. If already initialized it does nothing.
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

// Initialized will check if the Migrator is setup in this database.
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

// SetVersion updates the current version in the database.
func (sm SqliteMigrator) setVersion(version int) error {
	stmt := "UPDATE _migrator_ SET version = ?1"
	_, err := sm.db.Exec(stmt, version)
	return err
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

func (sm SqliteMigrator) Migrate() ([]Migration, error) {
	return sm.migrate(sm)
}
