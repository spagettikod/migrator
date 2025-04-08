package migrator

import (
	"database/sql"
	"errors"
	"log/slog"
)

type SqliteMigrator struct {
	db *sql.DB
}

func NewSqliteMigrator(db *sql.DB) SqliteMigrator {
	return SqliteMigrator{db: db}
}

// Init will set up the Migrator for the current database. If already initialized it does nothing.
func (sm SqliteMigrator) Init() error {
	initialized, err := sm.Initialized()
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
func (sm SqliteMigrator) Initialized() (bool, error) {
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

// Version returns the current version from the database.
func (sm SqliteMigrator) Version() (int, error) {
	initialized, err := sm.Initialized()
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

// SetVersion updates the current version in the database.
func (sm SqliteMigrator) SetVersion(version int) error {
	currentVersion, err := sm.Version()
	if err != nil {
		return err
	}
	stmt := "UPDATE _migrator_ SET version = ?1"
	slog.Debug("setting version", "currentVersion", currentVersion, "new_version", version, "sql", stmt)
	_, err = sm.db.Exec(stmt, version)
	return err
}

// Migrate will run the forward migrations in the array.
func (sm SqliteMigrator) Migrate(migrations []string) error {
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
