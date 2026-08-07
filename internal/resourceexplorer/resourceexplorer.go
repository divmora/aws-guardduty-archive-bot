package resourceexplorer

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
)

// GetAllEC2InstancesForAccountRegion fetches all EC2 instances from Resource Explorer for a specific account ID and region,
// and returns a map of instance IDs for O(1) local lookups.
func GetAllEC2InstancesForAccountRegion(ctx context.Context, client *resourceexplorer2.Client, viewArn string, accountId string, region string) (map[string]bool, error) {
	instanceCache := make(map[string]bool)

	// Fetch all EC2 instances for the given account and region
	query := fmt.Sprintf("resourcetype:ec2:instance account:%s region:%s", accountId, region)

	paginator := resourceexplorer2.NewSearchPaginator(client, &resourceexplorer2.SearchInput{
		QueryString: aws.String(query),
		ViewArn:     aws.String(viewArn),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to search resource explorer for account %s: %w", accountId, err)
		}

		for _, resource := range page.Resources {
			if resource.Arn != nil {
				// EC2 instance ARN format: arn:aws:ec2:region:account:instance/i-1234567890abcdef0
				parts := strings.Split(*resource.Arn, "/")
				if len(parts) > 1 {
					instanceId := parts[len(parts)-1]
					instanceCache[instanceId] = true
				}
			}
		}
	}

	return instanceCache, nil
}
