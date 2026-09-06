package config

import (
	"os"
	"reflect"
	"testing"
)

func TestParseRegions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single region",
			input:    "us-east-1",
			expected: []string{"us-east-1"},
		},
		{
			name:     "multiple regions with whitespace",
			input:    "us-east-1, us-west-2,  ap-south-1 ",
			expected: []string{"us-east-1", "us-west-2", "ap-south-1"},
		},
		{
			name:     "empty elements",
			input:    ",us-east-1,,us-west-2,",
			expected: []string{"us-east-1", "us-west-2"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRegions(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseRegions(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoadConfigWithOverrides(t *testing.T) {
	// Clear env vars for test isolation
	origRegions := os.Getenv("GUARDDUTY_REGIONS")
	origArn := os.Getenv("ORG_ALL_RESOURCES_VIEW_ARN")
	defer func() {
		os.Setenv("GUARDDUTY_REGIONS", origRegions)
		os.Setenv("ORG_ALL_RESOURCES_VIEW_ARN", origArn)
	}()

	t.Run("overrides take precedence", func(t *testing.T) {
		os.Setenv("GUARDDUTY_REGIONS", "env-region-1")
		os.Setenv("ORG_ALL_RESOURCES_VIEW_ARN", "arn:aws:view:env")

		cfg, err := LoadConfigWithOverrides("override-region-1,override-region-2", "arn:aws:view:override")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedRegions := []string{"override-region-1", "override-region-2"}
		if !reflect.DeepEqual(cfg.Regions, expectedRegions) {
			t.Errorf("cfg.Regions = %v, want %v", cfg.Regions, expectedRegions)
		}
		if cfg.OrgAllResourcesViewArn != "arn:aws:view:override" {
			t.Errorf("cfg.OrgAllResourcesViewArn = %q, want arn:aws:view:override", cfg.OrgAllResourcesViewArn)
		}
	})

	t.Run("falls back to environment variables", func(t *testing.T) {
		os.Setenv("GUARDDUTY_REGIONS", "fallback-region-1")
		os.Setenv("ORG_ALL_RESOURCES_VIEW_ARN", "arn:aws:view:fallback")

		cfg, err := LoadConfigWithOverrides("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedRegions := []string{"fallback-region-1"}
		if !reflect.DeepEqual(cfg.Regions, expectedRegions) {
			t.Errorf("cfg.Regions = %v, want %v", cfg.Regions, expectedRegions)
		}
		if cfg.OrgAllResourcesViewArn != "arn:aws:view:fallback" {
			t.Errorf("cfg.OrgAllResourcesViewArn = %q, want arn:aws:view:fallback", cfg.OrgAllResourcesViewArn)
		}
	})

	t.Run("errors on missing configuration", func(t *testing.T) {
		os.Unsetenv("GUARDDUTY_REGIONS")
		os.Unsetenv("ORG_ALL_RESOURCES_VIEW_ARN")

		_, err := LoadConfigWithOverrides("", "")
		if err == nil {
			t.Fatal("expected error on missing configuration, got nil")
		}
	})
}
