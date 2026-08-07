# AWS GuardDuty Archive Bot

AWS GuardDuty Archive Bot is a utility that automates the archiving of specific GuardDuty findings across multiple AWS regions based on configurable rules.

## Configuration

Configuration is managed via Environment Variables:

- `GUARDDUTY_REGIONS`: Comma-separated list of AWS regions to process (e.g. `ap-south-1,us-east-1,us-west-2`)
- `ORG_ALL_RESOURCES_VIEW_ARN`: The ARN for your AWS Resource Explorer View

## IAM Permissions

The Lambda function requires an execution role with the following least-privilege custom IAM policy. This grants permissions to read/archive GuardDuty findings and search Resource Explorer, alongside standard CloudWatch logging.

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "guardduty:ListDetectors",
                "guardduty:ListFindings",
                "guardduty:GetFindings",
                "guardduty:ArchiveFindings"
            ],
            "Resource": "arn:aws:guardduty:*:*:detector/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "resource-explorer-2:Search"
            ],
            "Resource": "arn:aws:resource-explorer-2:*:*:view/*/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "arn:aws:logs:*:*:log-group:/aws/lambda/*"
        }
    ]
}
```

## Running the Bot

First, ensure your AWS credentials are appropriately configured in your environment (e.g. via `~/.aws/credentials` or standard AWS environment variables) and any required environment variables are set in a `.env` file.

To run locally (simulating Lambda execution):

```bash
go run ./cmd/lambda
```

To build a binary:

```bash
make build
```
