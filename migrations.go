package migrator

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Migrations is a collection of Migrations, i.e. the YAML file.
type Migrations struct {
	Migrations []Migration `yaml:"migrations"`
}

func (ms Migrations) validate() error {
	for _, m := range ms.Migrations {
		if strings.TrimSpace(m.Up) == "" {
			return fmt.Errorf("migrator: \"up\"-statment for version %v is either missing or empty", m.Version())
		}
	}
	return nil
}

func (ms Migrations) enumerateMigrations() {
	for i := range ms.Migrations {
		ms.Migrations[i].version = i + 1
	}
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
	return migrations, migrations.validate()
}
