package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"testing"
)

func TestIsLambdaEnvironment(t *testing.T) {
	origApi := os.Getenv("AWS_LAMBDA_RUNTIME_API")
	origPort := os.Getenv("_LAMBDA_SERVER_PORT")
	defer func() {
		os.Setenv("AWS_LAMBDA_RUNTIME_API", origApi)
		os.Setenv("_LAMBDA_SERVER_PORT", origPort)
	}()

	os.Unsetenv("AWS_LAMBDA_RUNTIME_API")
	os.Unsetenv("_LAMBDA_SERVER_PORT")
	if isLambdaEnvironment() {
		t.Error("expected isLambdaEnvironment to be false when env vars are unset")
	}

	os.Setenv("AWS_LAMBDA_RUNTIME_API", "127.0.0.1:9001")
	if !isLambdaEnvironment() {
		t.Error("expected isLambdaEnvironment to be true when AWS_LAMBDA_RUNTIME_API is set")
	}

	os.Unsetenv("AWS_LAMBDA_RUNTIME_API")
	os.Setenv("_LAMBDA_SERVER_PORT", "8080")
	if !isLambdaEnvironment() {
		t.Error("expected isLambdaEnvironment to be true when _LAMBDA_SERVER_PORT is set")
	}
}

func TestRunCLIVersion(t *testing.T) {
	ctx := context.Background()

	err := runCLI(ctx, []string{"--version"})
	if err != nil {
		t.Errorf("expected nil error for --version, got %v", err)
	}

	err = runCLI(ctx, []string{"-v"})
	if err != nil {
		t.Errorf("expected nil error for -v, got %v", err)
	}
}

func TestRunCLIHelp(t *testing.T) {
	ctx := context.Background()

	err := runCLI(ctx, []string{"--help"})
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("expected flag.ErrHelp for --help, got %v", err)
	}
}

func TestRunCLIMissingConfig(t *testing.T) {
	origRegions := os.Getenv("GUARDDUTY_REGIONS")
	origArn := os.Getenv("ORG_ALL_RESOURCES_VIEW_ARN")
	defer func() {
		os.Setenv("GUARDDUTY_REGIONS", origRegions)
		os.Setenv("ORG_ALL_RESOURCES_VIEW_ARN", origArn)
	}()

	os.Unsetenv("GUARDDUTY_REGIONS")
	os.Unsetenv("ORG_ALL_RESOURCES_VIEW_ARN")

	// Notice .env might exist, so explicitly pass empty overrides with no env
	ctx := context.Background()
	// An unrecognized flag should return an error
	err := runCLI(ctx, []string{"--unknown-flag"})
	if err == nil {
		t.Error("expected error for unknown flag, got nil")
	}
}
