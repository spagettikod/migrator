package migrator

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Migrations is a collection of Migrations, i.e. the YAML file.
type Migrations struct {
	Migrations []Migration `yaml:"migrations"`
}

func (ms Migrations) enumerateMigrations() {
	for i := range ms.Migrations {
		ms.Migrations[i].version = i + 1
	}
}

// Migration represents an entry defined in the migration YAML.
type Migration struct {
	Comment string `yaml:"comment"`
	Up      string `yaml:"up"`
	Down    string `yaml:"down"`
	Err     error  `yaml:"-"`
	version int    `yaml:"-"`
}

func load() (Migrations, error) {
	filename, found := os.LookupEnv(envVarFile)
	if !found {
		return Migrations{}, ErrMigrationFileEnvMissing
	}
	b, err := os.ReadFile(filename)
	if err != nil {
		return Migrations{}, err
	}
	migrations := Migrations{}
	if err := yaml.Unmarshal(b, &migrations); err != nil {
		return Migrations{}, err
	}
	return migrations, nil
}

func (m Migration) stmt(dir direction) string {
	switch dir {
	case directionDown:
		return m.Down
	case directionUp:
		return m.Up
	}
	return ""
}
