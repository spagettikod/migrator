package migrator

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const postgresConnStr = "postgres://pgtest:testing@localhost:5432/migrator"

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION") != "1" {
		t.Skip("environment variable INTEGRATION is not set to 1, run testdata/integration_test.sh, skipping...")
	}
}

func TestPostgresUtil(t *testing.T) {
	skipIfNotIntegration(t)
	t.Run("Init", test_PostgresInit)
	t.Run("Version", test_PostgresVersion)
}

func connect(t *testing.T, target, filename string) *sql.DB {
	os.Setenv(envVarTarget, target)
	os.Setenv(envVarFile, filename)
	db, err := sql.Open("pgx", postgresConnStr)
	if err != nil {
		t.Fatalf("could not open database: %s", err)
	}
	return db
}

func tearDownTest(db *sql.DB) {
	os.Unsetenv(envVarFile)
	os.Unsetenv(envVarTarget)
	db.Close()
}

func test_PostgresInit(t *testing.T) {
	// Connect to the database
	db := connect(t, "2", "testdata/migrations.yml")
	defer tearDownTest(db)
	pm, err := NewPostgresMigrator(db, "")
	if err != nil {
		t.Fatalf("could not create PostgresMigrator: %s", err)
	}

	ok, err := pm.initialized()
	if err != nil {
		t.Fatalf("error while checking if initialized: %s", err)
	}
	if !ok {
		t.Errorf("expected database to be initialized ut it was not")
	}
}

func test_PostgresVersion(t *testing.T) {
	// Connect to the database
	db := connect(t, "2", "testdata/migrations.yml")
	defer tearDownTest(db)
	pm, err := NewPostgresMigrator(db, "")
	if err != nil {
		t.Fatalf("could not create PostgresMigrator: %s", err)
	}
	// initialize
	if err := pm.init(); err != nil {
		t.Fatalf("error while running init: %s", err)
	}

	// check version of newly initialized database
	version, err := pm.Version()
	if err != nil {
		t.Fatalf("error while running Version: %s", err)
	}
	if version != 0 {
		t.Fatalf("expected Version 0, but got %v", version)
	}

	// set version to 5
	if err := pm.setVersion(5); err != nil {
		t.Fatalf("failed while running SetVersion: %s", err)
	}

	// check version of database after update
	version, err = pm.Version()
	if err != nil {
		t.Fatalf("error while running Version: %s", err)
	}
	if version != 5 {
		t.Fatalf("expected Version 5, but got %v", version)
	}
}

func TestPostgresMigrate(t *testing.T) {
	skipIfNotIntegration(t)
	db := connect(t, "1", "testdata/postgres_migration_up.yml")
	defer tearDownTest(db)
	pm, err := NewPostgresMigrator(db, "")
	if err != nil {
		t.Fatalf("could not create PostgresMigrator: %s", err)
	}
	ran, err := pm.Migrate()
	if err != nil {
		t.Fatalf("error while running Upgrade: %s", err)
	}
	if len(ran) != 1 {
		t.Errorf("expected to have run 1 migration but ran %v", len(ran))
	}

	v, err := pm.Version()
	if err != nil {
		t.Fatalf("error while checking version: %s", err)
	}
	if v != 1 {
		t.Fatalf("expected version to be 1 after upgrade but was %v", v)
	}

	countStmt := "SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'test'"
	row := db.QueryRow(countStmt)
	count := -1
	if err := row.Scan(&count); err != nil {
		t.Fatalf("error while running verifying test: %s", err)
	}
	if count != 1 {
		t.Fatalf("expected to find table named 'test' but did not")
	}

	// downgrade
	os.Setenv(envVarTarget, "0")
	pm, err = NewPostgresMigrator(db, "")
	if err != nil {
		t.Fatalf("could not create PostgresMigrator: %s", err)
	}

	ran, err = pm.Migrate()
	if err != nil {
		t.Fatalf("error while running Upgrade: %s", err)
	}
	if len(ran) != 1 {
		t.Errorf("expected to have run 1 migration but ran %v", len(ran))
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
