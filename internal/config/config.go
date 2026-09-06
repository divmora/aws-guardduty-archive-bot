package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the configuration parameters required to run the bot.
type Config struct {
	Regions                []string
	OrgAllResourcesViewArn string
}

// parseRegions splits a comma-separated list of AWS regions, trimming any whitespace.
func parseRegions(raw string) []string {
	var regions []string
	for _, r := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(r)
		if trimmed != "" {
			regions = append(regions, trimmed)
		}
	}
	return regions
}

// loadDotEnv parses a simple .env file if it exists and populates missing environment variables.
func loadDotEnv(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			val = strings.Trim(val, `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

// LoadConfigWithOverrides reads configuration, giving precedence to CLI flag overrides
// before falling back to environment variables or a local .env file.
func LoadConfigWithOverrides(regionsOverride, viewArnOverride string) (*Config, error) {
	loadDotEnv(".env")

	var cfg Config

	regionsStr := regionsOverride
	if regionsStr == "" {
		regionsStr = os.Getenv("GUARDDUTY_REGIONS")
	}
	if regionsStr != "" {
		cfg.Regions = parseRegions(regionsStr)
	}

	cfg.OrgAllResourcesViewArn = viewArnOverride
	if cfg.OrgAllResourcesViewArn == "" {
		cfg.OrgAllResourcesViewArn = os.Getenv("ORG_ALL_RESOURCES_VIEW_ARN")
	}

	if len(cfg.Regions) == 0 || cfg.OrgAllResourcesViewArn == "" {
		return nil, fmt.Errorf("missing required configuration: regions (flag --regions or env GUARDDUTY_REGIONS) or view ARN (flag --view-arn or env ORG_ALL_RESOURCES_VIEW_ARN)")
	}

	return &cfg, nil
}

// LoadConfig reads configuration from environment variables or a local .env file.
func LoadConfig() (*Config, error) {
	return LoadConfigWithOverrides("", "")
}
