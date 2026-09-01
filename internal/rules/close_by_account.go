package rules

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/divmora/aws-guardduty-archive-bot/internal/guardduty"
	"github.com/divmora/aws-guardduty-archive-bot/internal/utils"

	awsGuardduty "github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
)

// ApplyCloseByAccountRule lists and archives all findings that belong to GuardDuty member accounts that are disabled.
func ApplyCloseByAccountRule(ctx context.Context, client *awsGuardduty.Client, detectorId string, approve bool) error {
	slog.Info("Applying Rule: Close findings by disabled target accounts...")

	accountsToClose, err := guardduty.GetDisabledAccounts(ctx, client, detectorId)
	if err != nil {
		return fmt.Errorf("failed to get disabled accounts: %w", err)
	}

	if len(accountsToClose) == 0 {
		slog.Info("No disabled member accounts found. Skipping rule.")
		return nil
	}

	slog.Info("Found disabled member accounts", "count", len(accountsToClose))

	// 1. Build criteria for finding these accounts (and ensure they are NOT already archived)
	criteria := &types.FindingCriteria{
		Criterion: map[string]types.Condition{
			"accountId": {
				Equals: accountsToClose,
			},
			"service.archived": {
				Equals: []string{"false"},
			},
		},
	}

	// 2. List findings
	findingIds, err := guardduty.ListFindings(ctx, client, detectorId, criteria, nil, 0)
	if err != nil {
		return fmt.Errorf("failed to list findings: %w", err)
	}

	if len(findingIds) == 0 {
		slog.Info("No findings found for the target accounts.")
		return nil
	}

	// 3. Fetch finding details
	findings, err := guardduty.GetFindingDetails(ctx, client, detectorId, findingIds)
	if err != nil {
		return fmt.Errorf("failed to get finding details: %w", err)
	}

	// 4. Double check that they actually belong to the targeted accounts
	var findingsToClose []string
	closedCountByAccount := make(map[string]int)

	for _, finding := range findings {
		accountId := "Unknown"
		if finding.AccountId != nil {
			accountId = *finding.AccountId
		}

		if utils.Contains(accountsToClose, accountId) {
			findingsToClose = append(findingsToClose, *finding.Id)
			closedCountByAccount[accountId]++
		}
	}

	// 5. Archive them
	if len(findingsToClose) > 0 {
		if approve {
			slog.Info("Archiving findings", "count", len(findingsToClose))
			err = guardduty.ArchiveFindings(ctx, client, detectorId, findingsToClose)
			if err != nil {
				return fmt.Errorf("failed to archive findings: %w", err)
			}
			slog.Info("Successfully closed (archived) the findings for this rule.")
		} else {
			slog.Info("DRY RUN: Findings for disabled accounts WOULD BE archived. Use --approve to execute.", "count", len(findingsToClose))
		}

		for acc, count := range closedCountByAccount {
			slog.Info("Close findings by disabled accounts summary", "accountId", acc, "closedCount", count)
		}
	} else {
		slog.Info("No findings to close after filtering.")
	}

	return nil
}
