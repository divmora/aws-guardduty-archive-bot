package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Regions                []string `yaml:"regions"`
	OrgAllResourcesViewArn string   `yaml:"org_all_resources_view_arn"`
}

// LoadConfig reads the YAML configuration file and parses it into a Config struct.
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return &cfg, nil
}
