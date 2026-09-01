package guardduty

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
)

// GetDetectorId retrieves the first available GuardDuty Detector ID in the current region.
func GetDetectorId(ctx context.Context, client *guardduty.Client) (string, error) {
	detectorsOutput, err := client.ListDetectors(ctx, &guardduty.ListDetectorsInput{})
	if err != nil {
		return "", err
	}

	if len(detectorsOutput.DetectorIds) == 0 {
		return "", fmt.Errorf("no GuardDuty detectors found in the current region")
	}
	return detectorsOutput.DetectorIds[0], nil
}

// ListFindings lists GuardDuty finding IDs based on the provided criteria, sorting, and limit.
func ListFindings(ctx context.Context, client *guardduty.Client, detectorId string, criteria *types.FindingCriteria, sortCriteria *types.SortCriteria, limit int) ([]string, error) {
	listFindingsInput := &guardduty.ListFindingsInput{
		DetectorId:      aws.String(detectorId),
		FindingCriteria: criteria,
		SortCriteria:    sortCriteria,
	}

	paginator := guardduty.NewListFindingsPaginator(client, listFindingsInput)
	var findingIds []string

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, id := range page.FindingIds {
			findingIds = append(findingIds, id)
			if limit > 0 && len(findingIds) >= limit {
				return findingIds, nil
			}
		}
	}
	return findingIds, nil
}

// GetDisabledAccounts fetches all GuardDuty members for a detector and returns the account IDs
// of those that are currently in a "Disabled" state.
func GetDisabledAccounts(ctx context.Context, client *guardduty.Client, detectorId string) ([]string, error) {
	var disabledAccounts []string

	paginator := guardduty.NewListMembersPaginator(client, &guardduty.ListMembersInput{
		DetectorId:     aws.String(detectorId),
		OnlyAssociated: aws.String("false"),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list members: %w", err)
		}

		for _, member := range page.Members {
			// Check relationship status. The exact string value might vary (e.g., "Disabled", "Removed", "Resigned")
			// but we primarily look for "Disabled".
			status := "Unknown"
			if member.RelationshipStatus != nil {
				status = *member.RelationshipStatus
			}

			if status == "Disabled" || status == "Removed" || status == "Resigned" {
				if member.AccountId != nil {
					disabledAccounts = append(disabledAccounts, *member.AccountId)
				}
			}
		}
	}

	return disabledAccounts, nil
}

// GetFindingDetails retrieves the full details for a given list of finding IDs, handling batch limits.
func GetFindingDetails(ctx context.Context, client *guardduty.Client, detectorId string, findingIds []string) ([]types.Finding, error) {
	var allFindings []types.Finding
	const batchSize = 50

	for i := 0; i < len(findingIds); i += batchSize {
		end := i + batchSize
		if end > len(findingIds) {
			end = len(findingIds)
		}
		batch := findingIds[i:end]

		getFindingsInput := &guardduty.GetFindingsInput{
			DetectorId: aws.String(detectorId),
			FindingIds: batch,
		}

		findingsOutput, err := client.GetFindings(ctx, getFindingsInput)
		if err != nil {
			return nil, err
		}

		allFindings = append(allFindings, findingsOutput.Findings...)
	}
	return allFindings, nil
}

// ArchiveFindings archives (closes) the provided GuardDuty findings, handling batch limits.
func ArchiveFindings(ctx context.Context, client *guardduty.Client, detectorId string, findingIds []string) error {
	const batchSize = 50

	for i := 0; i < len(findingIds); i += batchSize {
		end := i + batchSize
		if end > len(findingIds) {
			end = len(findingIds)
		}
		batch := findingIds[i:end]

		archiveInput := &guardduty.ArchiveFindingsInput{
			DetectorId: aws.String(detectorId),
			FindingIds: batch,
		}

		_, err := client.ArchiveFindings(ctx, archiveInput)
		if err != nil {
			return err
		}
	}
	return nil
}
