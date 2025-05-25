package migrator

import (
	"fmt"
	"os"
	"slices"
	"testing"
)

func TestTargetVersion(t *testing.T) {
	b := base{db: nil, migrations: Migrations{[]Migration{{}}}}
	// no target given should result in error
	_, err := b.parseTarget()
	if err == nil {
		t.Errorf("error was nil, expected an error")
	}
	if err != nil && err != ErrInvalidTargetVersion {
		t.Fatal(err)
	}

	os.Setenv(EnvVarTarget, "a")
	_, err = b.parseTarget()
	if err == nil {
		t.Errorf("error was nil, expected an error")
	}
	if err != nil && err != ErrInvalidTargetVersion {
		t.Fatal(err)
	}

	os.Setenv(EnvVarTarget, "-1")
	_, err = b.parseTarget()
	if err == nil {
		t.Errorf("error was nil, expected an error")
	}
	if err != nil && err != ErrInvalidTargetVersion {
		t.Fatal(err)
	}

	os.Setenv(EnvVarTarget, "5")
	_, err = b.parseTarget()
	if err == nil {
		t.Errorf("error was nil, expected an error")
	}
	if err != nil && err != ErrTargetOutOfBounds {
		t.Fatal(err)
	}

	os.Setenv(EnvVarTarget, "0")
	_, err = b.parseTarget()
	if err != nil {
		t.Errorf("undexpected error occured: %s", err)
	}
}
func TestValidTarget(t *testing.T) {
	type Case struct {
		base     base
		Target   int
		Expected bool
	}

	cases := []Case{
		{
			base:     base{migrations: Migrations{Migrations: []Migration{{}, {}}}},
			Target:   2,
			Expected: true,
		},
		{
			base:     base{migrations: Migrations{Migrations: []Migration{{}, {}}}},
			Target:   targetStart,
			Expected: true, // TargetStart is the starting point when no target has been run
		},
		{
			base:     base{migrations: Migrations{Migrations: []Migration{{}, {}}}},
			Target:   -1,
			Expected: false,
		},
	}

	for i, tc := range cases {
		if actual := tc.base.validTarget(tc.Target); actual != tc.Expected {
			t.Errorf("%v: expected %v but got %v", i, tc.Expected, actual)
		}
	}
}

func TestDirection(t *testing.T) {
	if migrationDirection(0, 2) != directionUp {
		t.Errorf("expected %v but got: %v", directionUp, migrationDirection(0, 2))
	}
	if migrationDirection(1, 0) != directionDown {
		t.Errorf("expected %v but got: %v", directionDown, migrationDirection(1, 0))
	}
	if migrationDirection(1, 1) != directionNone {
		t.Errorf("expected %v but got: %v", directionNone, migrationDirection(1, 1))
	}
}

func TestTargetMigrations(t *testing.T) {
	type Case struct {
		Migrations     Migrations
		CurrentVersion int
		Target         int
		Expected       []Migration
	}

	cases := []Case{
		{
			Migrations:     Migrations{Migrations: []Migration{{Up: "a"}, {Up: "b"}, {Up: "c"}}},
			CurrentVersion: targetStart,
			Target:         1,
			Expected:       []Migration{{Up: "a", version: 1}},
		},
		{
			Migrations:     Migrations{Migrations: []Migration{{Up: "a"}, {Up: "b"}, {Up: "c"}}},
			CurrentVersion: 3,
			Target:         1,
			Expected:       []Migration{{Up: "c", version: 3}, {Up: "b", version: 2}},
		},
		{
			Migrations:     Migrations{Migrations: []Migration{{Up: "a"}, {Up: "b"}, {Up: "c"}}},
			CurrentVersion: targetStart,
			Target:         3,
			Expected:       []Migration{{Up: "a", version: 1}, {Up: "b", version: 2}, {Up: "c", version: 3}},
		},
		{
			Migrations:     Migrations{Migrations: []Migration{{Up: "a"}, {Up: "b"}, {Up: "c"}}},
			CurrentVersion: 3,
			Target:         3,
			Expected:       []Migration{},
		},
	}

	for i, tc := range cases {
		os.Setenv(EnvVarTarget, fmt.Sprintf("%v", tc.Target))
		b, err := newBase(nil, tc.Migrations)
		if err != nil {
			t.Fatal(err)
		}
		actual := b.targetMigrations(tc.CurrentVersion)
		if !slices.Equal(actual, tc.Expected) {
			t.Errorf("%v: expected %v but got %v", i, tc.Expected, actual)
		}
	}
}
