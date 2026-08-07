# AWS GuardDuty Archive Bot

AWS GuardDuty Archive Bot is a utility that automates the archiving of specific GuardDuty findings across multiple AWS regions based on configurable rules.

## Configuration

Configuration is managed via `config.yaml`. Example configuration:

```yaml
regions:
  - "ap-south-1"
  - "us-east-1"
  - "us-west-2"
  - "ap-south-2"
  - "ap-southeast-1"

org_all_resources_view_arn: "arn:aws:resource-explorer-2:ap-south-1:123456789012:view/org-all-resources/..."
```

## Running the Bot

First, ensure your AWS credentials are appropriately configured in your environment (e.g. via `~/.aws/credentials` or standard AWS environment variables) and any required environment variables are set in a `.env` file.

To run the bot:

```bash
go run main.go -config config.yaml
```

To build a binary:

```bash
go build -o aws-guardduty-archive-bot
./aws-guardduty-archive-bot -config config.yaml
```
