# AWS GuardDuty Archive Bot

[![Latest Release](https://img.shields.io/github/v/release/divmora/aws-guardduty-archive-bot?logo=github)](https://github.com/divmora/aws-guardduty-archive-bot/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/divmora/aws-guardduty-archive-bot)](go.mod)
[![Documentation: DeepWiki](https://img.shields.io/badge/docs-DeepWiki-blue.svg)](https://deepwiki.com/divmora/aws-guardduty-archive-bot)
[![CI/CD](https://github.com/divmora/aws-guardduty-archive-bot/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/divmora/aws-guardduty-archive-bot/actions)
[![License: BSL 1.1](https://img.shields.io/badge/License-BSL_1.1-blue.svg)](https://github.com/divmora/.github/blob/main/LICENSING.md)
[![Security Policy](https://img.shields.io/badge/Security-Policy-green.svg)](SECURITY.md)

**AWS GuardDuty Archive Bot** is an automated AWS Lambda utility that cleans up stale, orphaned, and irrelevant Amazon GuardDuty security findings across multiple AWS regions and organization accounts based on configurable rules.

---

## Features & Archive Rules

The bot currently runs the following automated rules across configured AWS regions:

1. **Disabled Member Accounts Rule (`CloseByAccountRule`)**:
   - Discovers GuardDuty member accounts that have been disabled or disconnected.
   - Automatically archives any lingering active findings associated with those disabled accounts.

2. **Orphaned EC2 Instances Rule (`CheckOrphanInstancesRule`)**:
   - Queries active EC2 instance security findings that are older than 7 days.
   - Cross-references findings with the centralized **AWS Resource Explorer 2** view to check if the underlying EC2 instance has already been terminated.
   - Archives findings for non-existent (orphaned) instances.

---

## Execution Modes

The bot provides built-in safety controls via the invocation event payload:

- **Dry-Run Mode (Default)**:
  ```json
  { "approve": false }
  ```
  Scans findings, cross-references resources, and outputs structured audit logs without making any changes to GuardDuty.

- **Approve / Execution Mode**:
  ```json
  { "approve": true }
  ```
  Executes the archival action (`guardduty:ArchiveFindings`) for all qualifying findings.

---

## Configuration

Configuration is managed via Environment Variables:

| Variable | Description | Example |
| :--- | :--- | :--- |
| `GUARDDUTY_REGIONS` | Comma-separated list of AWS regions to process | `us-east-1,us-west-2,ap-south-1` |
| `ORG_ALL_RESOURCES_VIEW_ARN` | The ARN of the AWS Resource Explorer View for cross-account resource discovery | `arn:aws:resource-explorer-2:us-east-1:123456789012:view/all-resources/uuid` |

---

## IAM Permissions

The Lambda execution role requires the following least-privilege IAM policy to access GuardDuty, search Resource Explorer, and log to CloudWatch:

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
                "guardduty:GetMemberDetectors",
                "guardduty:ListMembers",
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

---

## Development & Building

### Prerequisites
- **Go 1.25+**
- **Make**
- **Docker** (optional, for container builds)

### Common Commands

```bash
# Setup development tools
make dev-setup

# Build local binary to bin/
make build

# Run unit tests
make test

# Format code
make fmt

# Run linter
make lint

# Package for AWS Lambda (generates lambda.zip)
make lambda-package
```

---

## Community & Contributing

- **[Contributing Guide](CONTRIBUTING.md)**: Review contribution guidelines, development workflows, and Conventional Commit requirements.
- **[Code of Conduct](CODE_OF_CONDUCT.md)**: Contributor Covenant Code of Conduct.
- **[Security Policy](SECURITY.md)**: Guidelines for reporting security vulnerabilities responsibly.

---

## License & Commercial Use

This project is licensed under the **Business Source License 1.1 (BSL 1.1)**.

- **Non-Production & Evaluation:** Free to use, modify, and test in non-production environments (local development, staging, QA, CI/CD pipelines, and proof-of-concept evaluation).
- **Production & Commercial Use:** Deploying or executing in production environments, embedding into commercial products, or offering as a managed service requires a commercial license (EULA) from **DIVMORA Technologies**.
- **Open Source Transition:** Each release automatically converts to the **Apache License, Version 2.0** three (3) years after its release date.

For commercial licensing inquiries, enterprise support, or questions, please contact **licensing@divmora.com** or visit [divmora.com](https://divmora.com).


