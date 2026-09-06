package runner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/divmora/aws-guardduty-archive-bot/internal/config"
	"github.com/divmora/aws-guardduty-archive-bot/internal/guardduty"
	"github.com/divmora/aws-guardduty-archive-bot/internal/rules"
	"github.com/divmora/aws-guardduty-archive-bot/internal/utils"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awsGuardduty "github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
)

// Run executes the GuardDuty archival rules across all configured regions.
// When approve is false, findings qualifying for archival are logged but not modified (dry-run).
// When approve is true, qualifying findings are archived via GuardDuty API.
func Run(ctx context.Context, approve bool, cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if approve {
		slog.Warn("Running in APPROVE mode. Findings will be archived.")
	} else {
		slog.Info("Running in DRY-RUN mode. No findings will be archived.")
	}

	// Parse region from the Resource Explorer View ARN to create the reClient
	parsedArn, err := utils.ParseArn(cfg.OrgAllResourcesViewArn)
	if err != nil {
		slog.Error("failed to parse org_all_resources_view_arn", "error", err)
		return fmt.Errorf("failed to parse org_all_resources_view_arn: %w", err)
	}
	reRegion := parsedArn.Region

	reAwsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(reRegion))
	if err != nil {
		slog.Error("unable to load SDK config for Resource Explorer", "region", reRegion, "error", err)
		return fmt.Errorf("unable to load SDK config for Resource Explorer (%s): %w", reRegion, err)
	}
	reClient := resourceexplorer2.NewFromConfig(reAwsCfg)

	// Loop over each GuardDuty region specified in the config
	for _, region := range cfg.Regions {
		// Check for context cancellation before processing next region
		if err := ctx.Err(); err != nil {
			slog.Warn("execution stopped due to context cancellation", "error", err)
			return err
		}

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
	return nil
}
