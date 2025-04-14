package migrator

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const postgresConnStr = "host=localhost port=5432 user=pgtest password=testing dbname=migrator"

func TestPostgresInitialized(t *testing.T) {
	// Connect to the database
	db, err := sql.Open("pgx", postgresConnStr)
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}
	defer db.Close()

	// Ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		t.Fatalf("could not ping database: %s", err)
		return
	}

	fmt.Println("Connected to PostgreSQL database!")
	sm := NewPostgresMigrator(db, "")
	init, err := sm.Initialized()
	if err != nil {
		t.Fatalf("error while running Initialized: %s", err)
	}
	if init {
		t.Fatal("expected Initlized to return false but it returned true")
	}
}

func TestPostgresInit(t *testing.T) {
	// Connect to the database
	db, err := sql.Open("pgx", postgresConnStr)
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}
	defer db.Close()
	sm := NewPostgresMigrator(db, "")

	// should not be initialized
	isInitialized, err := sm.Initialized()
	if err != nil {
		t.Fatalf("error while running Initialized: %s", err)
	}
	if isInitialized {
		t.Fatal("expected Initialized to return false but it returned true")
	}

	// initialize
	if err := sm.Init(); err != nil {
		t.Fatalf("error while running Init: %s", err)
	}

	// should now be initialized
	isInitialized, err = sm.Initialized()
	if err != nil {
		t.Fatalf("error while running Initialized: %s", err)
	}
	if !isInitialized {
		t.Fatal("expected Initialized to return true but it returned false")
	}
}

func TestPostgresVersion(t *testing.T) {
	db, err := sql.Open("pgx", postgresConnStr)
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}
	defer db.Close()
	sm := NewPostgresMigrator(db, "")
	// initialize
	if err := sm.Init(); err != nil {
		t.Fatalf("error while running Init: %s", err)
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
	if err := sm.SetVersion(5); err != nil {
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

func TestPostgresMigrate(t *testing.T) {
	db, err := sql.Open("pgx", postgresConnStr)
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}
	defer db.Close()
	sm := NewPostgresMigrator(db, "")

	// initialize
	if err := sm.Init(); err != nil {
		t.Fatalf("error while running Init: %s", err)
	}

	migrations := []string{
		"CREATE TABLE test (id INTEGER PRIMARY KEY)",
	}

	if err := sm.Migrate(migrations); err != nil {
		t.Fatalf("error while running Upgrade: %s", err)
	}

	v, err := sm.Version()
	if err != nil {
		t.Fatalf("error while checking version: %s", err)
	}
	if v != 1 {
		t.Fatalf("expected version to be 1 after upgrade but was %v", v)
	}

	sql := "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'test'"
	rows, err := sm.db.Query(sql)
	if err != nil {
		t.Fatalf("error while running verifying test: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		return
	}
	t.Fatalf("expected to find table named 'test' but did not")
}
