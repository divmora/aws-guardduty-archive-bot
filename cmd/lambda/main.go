package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/divmora/aws-guardduty-archive-bot/internal/config"
	"github.com/divmora/aws-guardduty-archive-bot/internal/runner"
)

// Event defines the payload for the Lambda invocation.
type Event struct {
	Approve bool `json:"approve"`
}

// Version holds the application version, injected at build time via -ldflags.
var Version = "dev"

// isLambdaEnvironment detects if the process is running within an AWS Lambda environment.
func isLambdaEnvironment() bool {
	return os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" || os.Getenv("_LAMBDA_SERVER_PORT") != ""
}

// HandleRequest is the entrypoint handler for AWS Lambda invocations.
func HandleRequest(ctx context.Context, event Event) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting AWS GuardDuty Archive Bot", "version", Version, "mode", "lambda")

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}

	return runner.Run(ctx, event.Approve, cfg)
}

// runCLI parses command-line arguments and runs the bot in standalone CLI / container mode.
func runCLI(ctx context.Context, args []string) error {
	flagSet := flag.NewFlagSet("aws-guardduty-archive-bot", flag.ContinueOnError)

	approve := flagSet.Bool("approve", false, "Execute archival mutations (default false: dry-run mode)")
	regions := flagSet.String("regions", "", "Comma-separated list of AWS regions (overrides GUARDDUTY_REGIONS)")
	viewArn := flagSet.String("view-arn", "", "Resource Explorer View ARN (overrides ORG_ALL_RESOURCES_VIEW_ARN)")
	jsonLogs := flagSet.Bool("json", false, "Output logs in JSON format instead of readable text")
	showVersion := flagSet.Bool("version", false, "Print version and exit")
	flagSet.BoolVar(showVersion, "v", false, "Print version and exit (shorthand)")

	flagSet.Usage = func() {
		fmt.Fprintf(flagSet.Output(), "AWS GuardDuty Archive Bot (%s)\n\n", Version)
		fmt.Fprintf(flagSet.Output(), "Usage:\n  aws-guardduty-archive-bot [flags]\n\nFlags:\n")
		flagSet.PrintDefaults()
		fmt.Fprintf(flagSet.Output(), "\nEnvironment Variables:\n")
		fmt.Fprintf(flagSet.Output(), "  GUARDDUTY_REGIONS          Comma-separated list of AWS regions\n")
		fmt.Fprintf(flagSet.Output(), "  ORG_ALL_RESOURCES_VIEW_ARN Resource Explorer View ARN\n")
	}

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if *showVersion {
		fmt.Printf("aws-guardduty-archive-bot version %s\n", Version)
		return nil
	}

	// Configure structured logger
	var handler slog.Handler
	if *jsonLogs {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler))

	slog.Info("Starting AWS GuardDuty Archive Bot", "version", Version, "mode", "cli")

	cfg, err := config.LoadConfigWithOverrides(*regions, *viewArn)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return err
	}

	return runner.Run(ctx, *approve, cfg)
}

func main() {
	if isLambdaEnvironment() {
		lambda.Start(HandleRequest)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runCLI(ctx, os.Args[1:]); err != nil {
		if !errors.Is(err, flag.ErrHelp) {
			slog.Error("CLI execution failed", "error", err)
			os.Exit(1)
		}
	}
}
