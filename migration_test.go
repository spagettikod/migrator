package migrator

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	os.Setenv(envVarFile, "testdata/migrations.yml")
	defer os.Unsetenv(envVarFile)
	m, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if m.Migrations[0].Comment != "My first migration" {
		t.Errorf("expected %s but got %s", "My first migration", m.Migrations[0].Comment)
	}
	if strings.TrimSpace(m.Migrations[1].Up) != "DROP TABLE test" {
		t.Errorf("expected %s but got %s", "DROP TABLE test", m.Migrations[1].Up)
	}
	if strings.TrimSpace(m.Migrations[1].Down) != "CREATE TABLE test (id INTEGER PRIMARY KEY)" {
		t.Errorf("expected %s but got %s", "DROP TABLE test", m.Migrations[1].Down)
	}
}
