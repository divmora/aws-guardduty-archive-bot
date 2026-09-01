package rules

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/divmora/aws-guardduty-archive-bot/internal/guardduty"
	"github.com/divmora/aws-guardduty-archive-bot/internal/resourceexplorer"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsGuardduty "github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
)

// CheckOrphanInstancesRule checks active EC2 instance findings against the Resource Explorer
// cache to find instances that no longer exist, and archives them if they are older than 7 days.
func CheckOrphanInstancesRule(ctx context.Context, gdClient *awsGuardduty.Client, reClient *resourceexplorer2.Client, detectorId string, viewArn string, approve bool) error {
	slog.Info("Applying Rule: Check for orphaned EC2 instances...")

	activeInstances := make(map[string]bool)

	// 2. Fetch all active findings for EC2 instances
	criteria := &types.FindingCriteria{
		Criterion: map[string]types.Condition{
			"resource.resourceType": {
				Equals: []string{"Instance"},
			},
			"service.archived": {
				Equals: []string{"false"},
			},
		},
	}

	sortCriteria := &types.SortCriteria{
		AttributeName: aws.String("updatedAt"),
		OrderBy:       types.OrderByAsc,
	}

	findingIds, err := guardduty.ListFindings(ctx, gdClient, detectorId, criteria, sortCriteria, 0)
	if err != nil {
		return fmt.Errorf("failed to list findings: %w", err)
	}

	if len(findingIds) == 0 {
		slog.Info("No active EC2 instance findings found.")
		return nil
	}

	// 3. Fetch finding details
	findings, err := guardduty.GetFindingDetails(ctx, gdClient, detectorId, findingIds)
	if err != nil {
		return fmt.Errorf("failed to get finding details: %w", err)
	}

	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	var oldFindings []types.Finding
	for _, finding := range findings {
		if finding.UpdatedAt != nil {
			updatedAt, err := time.Parse(time.RFC3339, *finding.UpdatedAt)
			if err == nil && updatedAt.Before(sevenDaysAgo) {
				oldFindings = append(oldFindings, finding)
			}
		}
	}
	findings = oldFindings

	if len(findings) == 0 {
		slog.Info("No active EC2 instance findings older than 7 days found.")
		return nil
	}

	slog.Info("Found active EC2 instance findings older than 7 days", "count", len(findings))

	// Pre-fetch EC2 instances per account/region combination found in findings
	type accountRegion struct {
		account string
		region  string
	}
	uniqueCombos := make(map[accountRegion]bool)
	for _, finding := range findings {
		if finding.AccountId != nil && finding.Region != nil {
			combo := accountRegion{account: *finding.AccountId, region: *finding.Region}
			uniqueCombos[combo] = true
		}
	}

	slog.Info("Fetching EC2 instances from Resource Explorer", "uniqueAccountRegions", len(uniqueCombos))
	for combo := range uniqueCombos {
		instancesForAccReg, err := resourceexplorer.GetAllEC2InstancesForAccountRegion(ctx, reClient, viewArn, combo.account, combo.region)
		if err != nil {
			slog.Warn("failed to fetch instances", "account", combo.account, "region", combo.region, "error", err)
			continue
		}
		for instId := range instancesForAccReg {
			activeInstances[instId] = true
		}
	}
	slog.Info("Cached active EC2 instances", "total", len(activeInstances))

	// 4. Check findings against local cache
	var findingsToArchive []string
	summary := make(map[string]map[string]int)

	for _, finding := range findings {
		accountId := "Unknown"
		if finding.AccountId != nil {
			accountId = *finding.AccountId
		}

		region := "Unknown"
		if finding.Region != nil {
			region = *finding.Region
		}

		instanceId := "Unknown"
		if finding.Resource != nil && finding.Resource.InstanceDetails != nil && finding.Resource.InstanceDetails.InstanceId != nil {
			instanceId = *finding.Resource.InstanceDetails.InstanceId
		}

		if instanceId == "Unknown" {
			continue
		}

		// Local O(1) lookup
		isMissing := !activeInstances[instanceId]

		if isMissing {
			findingsToArchive = append(findingsToArchive, *finding.Id)

			// Update summary map
			if summary[accountId] == nil {
				summary[accountId] = make(map[string]int)
			}
			summary[accountId][region]++
		}
	}

	// 5. Archive missing instances findings and print summary
	if len(findingsToArchive) > 0 {
		if approve {
			slog.Info("Archiving findings for missing instances", "count", len(findingsToArchive))
			err = guardduty.ArchiveFindings(ctx, gdClient, detectorId, findingsToArchive)
			if err != nil {
				return fmt.Errorf("failed to archive findings: %w", err)
			}
		} else {
			slog.Info("DRY RUN: Findings for missing instances WOULD BE archived. Use --approve to execute.", "count", len(findingsToArchive))
		}

		for accId, regions := range summary {
			for reg, count := range regions {
				action := "archived"
				if !approve {
					action = "ready to archive"
				}
				slog.Info("Orphaned instances summary", "accountId", accId, "region", reg, "count", count, "status", action)
			}
		}
	} else {
		slog.Info("No findings needed to be archived for orphaned instances.")
	}

	return nil
}
