package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

type PostgresMigrator struct {
	db     *sql.DB
	schema string
}

func NewPostgresMigrator(db *sql.DB, schema string) PostgresMigrator {
	if schema == "" {
		schema = "public"
	}
	return PostgresMigrator{db: db, schema: schema}
}

// Init will set up the Migrator for the current database. If already initialized it does nothing.
func (sm PostgresMigrator) Init() error {
	initialized, err := sm.Initialized()
	if err != nil {
		return err
	}
	if !initialized {
		_, err = sm.db.Exec(fmt.Sprintf("CREATE TABLE %s._migrator_ (version INTEGER NOT NULL) STRICT", sm.schema))
		if err != nil {
			return err
		}
		_, err = sm.db.Exec(fmt.Sprintf("INSERT INTO %s._migrator_ (version) VALUES (0)", sm.schema))
	}
	return err
}

// Initialized will check if the Migrator is setup in this database.
func (sm PostgresMigrator) Initialized() (bool, error) {
	row := sm.db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_schema = ?1 AND table_name = 'users'", sm.schema)
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

// Version returns the current version from the database.
func (sm PostgresMigrator) Version() (int, error) {
	initialized, err := sm.Initialized()
	if err != nil {
		return -1, err
	}
	if !initialized {
		return 0, ErrMigratorNotInitialized
	}

	row := sm.db.QueryRow(fmt.Sprintf("SELECT version FROM %s._migrator_", sm.schema))
	version := 0
	if err := row.Scan(&version); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return -1, err
	}
	return version, nil
}

// SetVersion updates the current version in the database.
func (sm PostgresMigrator) SetVersion(version int) error {
	currentVersion, err := sm.Version()
	if err != nil {
		return err
	}
	stmt := fmt.Sprintf("UPDATE %s._migrator_ SET version = ?1", sm.schema)
	slog.Debug("setting version", "currentVersion", currentVersion, "new_version", version, "sql", stmt)
	_, err = sm.db.Exec(stmt, version)
	return err
}

// Migrate will run the forward migrations in the array.
func (sm PostgresMigrator) Migrate(migrations []string) error {
	v, err := sm.Version()
	if err != nil {
		return err
	}
	// if no migrations have run we're at version -1, to kickstart migrations we must start at v==0
	if v == -1 {
		v = 0
	}
	slog.Debug("migration check", "current_migration", v, "available_migrations", len(migrations)-v)
	for i := v; i < len(migrations); i++ {
		slog.Debug("migrating", "from", i, "to", i+1, "migration", migrations[i])
		if _, err := sm.db.Exec(migrations[i]); err != nil {
			return err
		}
		sm.SetVersion(i + 1)
	}
	return nil
}
