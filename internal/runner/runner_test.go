package runner

import (
	"context"
	"testing"

	"github.com/divmora/aws-guardduty-archive-bot/internal/config"
)

func TestRunNilConfig(t *testing.T) {
	ctx := context.Background()
	err := Run(ctx, false, nil)
	if err == nil {
		t.Fatal("expected error when config is nil, got nil")
	}
}

func TestRunInvalidArn(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		Regions:                []string{"us-east-1"},
		OrgAllResourcesViewArn: "not-an-arn",
	}

	err := Run(ctx, false, cfg)
	if err == nil {
		t.Fatal("expected error with invalid ARN, got nil")
	}
}

func TestRunCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	cfg := &config.Config{
		Regions:                []string{"us-east-1"},
		OrgAllResourcesViewArn: "arn:aws:resource-explorer-2:us-east-1:123456789012:view/all-resources/uuid",
	}

	err := Run(ctx, false, cfg)
	if err == nil {
		t.Fatal("expected error with canceled context, got nil")
	}
}
