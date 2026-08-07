package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Regions                []string
	OrgAllResourcesViewArn string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() (*Config, error) {
	var cfg Config

	regionsStr := os.Getenv("GUARDDUTY_REGIONS")
	if regionsStr != "" {
		cfg.Regions = strings.Split(regionsStr, ",")
	}
	cfg.OrgAllResourcesViewArn = os.Getenv("ORG_ALL_RESOURCES_VIEW_ARN")

	if len(cfg.Regions) == 0 || cfg.OrgAllResourcesViewArn == "" {
		return nil, fmt.Errorf("missing required env vars GUARDDUTY_REGIONS or ORG_ALL_RESOURCES_VIEW_ARN")
	}

	return &cfg, nil
}
