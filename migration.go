package migrator

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Migrations struct {
	Migrations []Migration `yaml:"migrations"`
}

func (ms Migrations) enumerateMigrations() {
	for i := range ms.Migrations {
		ms.Migrations[i].version = i + 1
	}
}

type Migration struct {
	Comment string `yaml:"comment"`
	Up      string `yaml:"up"`
	Down    string `yaml:"down"`
	Err     error  `yaml:"-"`
	version int    `yaml:"-"`
}

func Load() (Migrations, error) {
	filename, found := os.LookupEnv(EnvVarFile)
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

func (m Migration) Stmt(dir Direction) string {
	switch dir {
	case DirectionDown:
		return m.Down
	case DirectionUp:
		return m.Up
	}
	return ""
}
