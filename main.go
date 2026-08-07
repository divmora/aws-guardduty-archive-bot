package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/divmora/aws-guardduty-archive-bot/internal/config"
	"github.com/divmora/aws-guardduty-archive-bot/internal/guardduty"
	"github.com/divmora/aws-guardduty-archive-bot/internal/rules"
	"github.com/divmora/aws-guardduty-archive-bot/internal/utils"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsGuardduty "github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/joho/godotenv"
)

func main() {
	// Configure structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Define and parse CLI flags
	approvePtr := flag.Bool("approve", false, "Set this flag to actually archive findings. Without it, the script runs in dry-run mode and only prints a summary.")
	flag.Parse()
	approve := *approvePtr

	if approve {
		slog.Warn("Running in APPROVE mode. Findings will be archived.")
	} else {
		slog.Info("Running in DRY-RUN mode. No findings will be archived. Use --approve to execute.")
	}

	// Load .env file if it exists
	_ = godotenv.Load()
	ctx := context.TODO()

	// Load configuration from YAML
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	if len(cfg.Regions) == 0 {
		slog.Error("no regions specified in config.yaml")
		os.Exit(1)
	}

	// Parse region from the Resource Explorer View ARN to create the reClient
	parsedArn, err := utils.ParseArn(cfg.OrgAllResourcesViewArn)
	if err != nil {
		slog.Error("failed to parse org_all_resources_view_arn", "error", err)
		os.Exit(1)
	}
	reRegion := parsedArn.Region

	reAwsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(reRegion))
	if err != nil {
		slog.Error("unable to load SDK config for Resource Explorer", "region", reRegion, "error", err)
		os.Exit(1)
	}
	reClient := resourceexplorer2.NewFromConfig(reAwsCfg)

	// Loop over each GuardDuty region specified in the config
	for _, region := range cfg.Regions {
		slog.Info("Processing Region", "region", region)

		// Load AWS configuration for the specific region
		awsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region))
		if err != nil {
			slog.Error("unable to load SDK config", "region", region, "error", err)
			continue
		}

		gdClient := awsGuardduty.NewFromConfig(awsCfg)

		// Get Detector ID for the region
		detectorId, err := guardduty.GetDetectorId(ctx, gdClient)
		if err != nil {
			slog.Error("failed to get detector ID", "region", region, "error", err)
			continue
		}
		slog.Info("Using Detector ID", "detectorId", detectorId)

		// --- Rule Engine Execution ---

		// 1st Rule: Close all findings for disabled member accounts
		slog.Info("Executing CloseByAccountRule", "region", region)
		err = rules.ApplyCloseByAccountRule(ctx, gdClient, detectorId, approve)
		if err != nil {
			slog.Error("Error applying CloseByAccountRule", "region", region, "error", err)
		}

		// 2nd Rule: Find EC2 instances that no longer exist and print their details
		slog.Info("Executing CheckOrphanInstancesRule", "region", region)
		err = rules.CheckOrphanInstancesRule(ctx, gdClient, reClient, detectorId, cfg.OrgAllResourcesViewArn, approve)
		if err != nil {
			slog.Error("Error applying CheckOrphanInstancesRule", "region", region, "error", err)
		}
	}

	slog.Info("All rules executed successfully across all configured regions.")
}
