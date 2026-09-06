# Agent Guidelines for AWS GuardDuty Archive Bot

Welcome to the workspace-specific agent guidelines. These instructions apply to AI agents and developers working on the `aws-guardduty-archive-bot` codebase.

---

## 1. Project Architecture & Layout

This project follows standard Go conventions for serverless AWS Lambda tooling:

- `cmd/lambda/main.go`: The dual-mode entrypoint handler. Automatically detects whether it is running as an AWS Lambda function (`HandleRequest`) or a standalone CLI / container process (`runCLI`). Configures structured logging and delegates execution to `internal/runner`.
- `internal/runner/`: The core rule execution engine (`runner.Run`). Manages client initialization per region, context cancellation, and sequentially dispatches execution to individual rules.
- `internal/config/`: Configuration loading and validation from environment variables (`GUARDDUTY_REGIONS`, `ORG_ALL_RESOURCES_VIEW_ARN`) with support for CLI flag overrides and local `.env` files.
- `internal/guardduty/`: AWS GuardDuty SDK v2 wrappers (Detector lookup, finding search/pagination, member account status, finding archival).
- `internal/resourceexplorer/`: AWS Resource Explorer 2 SDK v2 client queries for cross-account resource cache lookups (e.g. EC2 instance status).
- `internal/rules/`: Pluggable archival rule logic (e.g. `CloseByAccountRule`, `CheckOrphanInstancesRule`).
- `internal/utils/`: Shared utilities such as ARN parsing and slice helpers.

---

## 2. Core Engineering Principles

### Safety & Dry-Run Guarantee
- **Mandatory Dry-Run Support**: Every rule function in `internal/rules/` **MUST** accept an `approve bool` parameter.
- When `approve == false`, the rule must query, filter, evaluate, and log the findings that qualify for archival, but **MUST NOT** invoke `guardduty:ArchiveFindings`.
- Only when `approve == true` should write/archive mutations be executed against AWS.

### Structured Logging (`log/slog`)
- Use Go standard library `log/slog` for all log outputs.
- Prefer structured key-value pairs (e.g., `slog.Info("Processing Region", "region", region, "count", count)`) instead of string formatting with `fmt.Sprintf`.
- Use appropriate log levels: `slog.Info` for milestones, `slog.Warn` for non-fatal skipped items or warnings, and `slog.Error` for operation failures.

### AWS SDK v2 Best Practices
- Always propagate `context.Context` through all SDK calls and rule functions.
- Avoid global AWS SDK clients; initialize clients with explicit region configurations (`awsConfig.WithRegion(region)`).
- Handle API pagination properly when querying GuardDuty or Resource Explorer to avoid silent truncation on large environments.
- Maintain least-privilege IAM requirements and document any new IAM action additions in `README.md`.

---

## 3. Contribution & Release Standards

### Conventional Commits
All commits must strictly follow the Conventional Commits specification so that [Release Please](https://github.com/googleapis/release-please) can correctly calculate semantic version bumps and generate release notes:
- `feat:` for new features or rules (bumps minor version).
- `fix:` for bug fixes (bumps patch version).
- `docs:` for documentation updates.
- `refactor:` for code refactoring without behavioral changes.
- `test:` for test additions or updates.
- `chore:` for tooling, dependencies, or configuration changes.

### Code Verification
Before submitting or finalizing changes, ensure:
```bash
make fmt   # Format code
make lint  # Run golangci-lint
make test  # Execute test suite
make build # Verify compilation to bin/
```

### Documentation Integrity
- When modifying rules, configuration options, or IAM requirements, update `README.md` to keep documentation accurate.
- Maintain license headers and respect the project's **Business Source License 1.1 (BSL 1.1)** terms.


