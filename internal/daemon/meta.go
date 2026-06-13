package daemon

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Meta stores runtime information about a daemon session.
type Meta struct {
	Profile   string    `yaml:"profile"`
	DBType    string    `yaml:"db_type"`
	Host      string    `yaml:"host"`
	StartTime time.Time `yaml:"start_time"`
}

// SaveMeta writes daemon metadata to a file.
func SaveMeta(path string, m *Meta) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadMeta reads daemon metadata from a file.
func LoadMeta(path string) (*Meta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Meta
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
