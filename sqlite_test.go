package migrator

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func initSQLiteTest(t *testing.T) *sql.DB {
	os.Unsetenv(EnvVarFile)
	os.Unsetenv(EnvVarTarget)
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}
	return db
}

func TestSQLiteInitialized(t *testing.T) {
	db := initSQLiteTest(t)
	os.Setenv(EnvVarTarget, "2")
	os.Setenv(EnvVarFile, "testdata/migrations.yml")
	defer os.Unsetenv(EnvVarFile)
	sm, err := NewSqliteMigrator(db)
	if err != nil {
		t.Fatalf("error while creating migrator: %s", err)
	}
	ok, err := sm.initialized()
	if err != nil {
		t.Fatalf("error while checking if initialized: %s", err)
	}
	if !ok {
		t.Errorf("expected database to be initialized ut it was not")
	}
}

func TestSQLiteVersion(t *testing.T) {
	db := initSQLiteTest(t)
	os.Setenv(EnvVarTarget, "2")
	os.Setenv(EnvVarFile, "testdata/migrations.yml")
	defer os.Unsetenv(EnvVarFile)
	sm, err := NewSqliteMigrator(db)
	if err != nil {
		t.Fatalf("error while creating migrator: %s", err)
	}

	// check version of newly initialized database
	version, err := sm.Version()
	if err != nil {
		t.Fatalf("error while running Version: %s", err)
	}
	if version != 0 {
		t.Fatalf("expected Version 0, but got %v", version)
	}

	// set version to 5
	if err := sm.setVersion(5); err != nil {
		t.Fatalf("failed while running SetVersion: %s", err)
	}

	// check version of database after update
	version, err = sm.Version()
	if err != nil {
		t.Fatalf("error while running Version: %s", err)
	}
	if version != 5 {
		t.Fatalf("expected Version 5, but got %v", version)
	}
}

func TestSQLiteMigrate(t *testing.T) {
	db := initSQLiteTest(t)
	os.Setenv(EnvVarTarget, "1")
	os.Setenv(EnvVarFile, "testdata/sqlite_migration_up.yml")
	defer os.Unsetenv(EnvVarFile)
	defer db.Close()
	sm, err := NewSqliteMigrator(db)
	if err != nil {
		t.Fatalf("could not load migration yaml: %s", err)
	}

	ran, err := sm.Migrate()
	if err != nil {
		t.Fatalf("error while running Upgrade: %s", err)
	}
	if len(ran) != 1 {
		t.Errorf("expected to have run 1 migration but ran %v", len(ran))
	}

	v, err := sm.Version()
	if err != nil {
		t.Fatalf("error while checking version: %s", err)
	}
	if v != 1 {
		t.Fatalf("expected version to be 1 after upgrade but was %v", v)
	}

	countStmt := "SELECT COUNT(1)	FROM sqlite_master WHERE type='table' AND name='test'"
	row := db.QueryRow(countStmt)
	count := -1
	if err := row.Scan(&count); err != nil {
		t.Fatalf("error while running verifying test: %s", err)
	}
	if count != 1 {
		t.Fatalf("expected to find table named 'test' but did not")
	}

	// downgrade
	os.Setenv(EnvVarTarget, "0")
	sm, err = NewSqliteMigrator(db)
	if err != nil {
		t.Fatalf("could not load migration yaml: %s", err)
	}
	ran, err = sm.Migrate()
	if err != nil {
		t.Fatalf("error while running Upgrade: %s", err)
	}
	if len(ran) != 1 {
		t.Errorf("expected to have run 1 migration but ran %v", len(ran))
	}

	v, err = sm.Version()
	if err != nil {
		t.Fatalf("error while checking version: %s", err)
	}
	if v != 0 {
		t.Fatalf("expected version to be 0 after upgrade but was %v", v)
	}

	row = db.QueryRow(countStmt)
	count = -1
	if err := row.Scan(&count); err != nil {
		t.Fatalf("error while running verifying test: %s", err)
	}
	if count != 0 {
		t.Fatalf("didn't expect to find table named 'test' but did")
	}
}
