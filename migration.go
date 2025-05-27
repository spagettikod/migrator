package migrator

// Migration represents an entry defined in the migration YAML.
type Migration struct {
	Comment string `yaml:"comment"`
	Up      string `yaml:"up"`
	Down    string `yaml:"down"`
	Err     error  `yaml:"-"`
	version int    `yaml:"-"`
}

// Version return the version number given for this migration. A migration gets
// it version from its position in the migrations YAML-file.
func (m Migration) Version() int {
	return m.version
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
