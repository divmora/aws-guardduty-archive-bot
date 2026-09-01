# Contributing to AWS GuardDuty Archive Bot

Thank you for your interest in contributing to **AWS GuardDuty Archive Bot**! We welcome bug reports, documentation improvements, feature suggestions, and code contributions.

Please review this guide before submitting issues or pull requests.

---

## Code of Conduct

This project adheres to the Contributor Covenant [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

---

## Getting Started

### Prerequisites

Ensure you have the following installed on your development machine:
- **Go 1.25+**: [golang.org](https://golang.org/dl/)
- **Make**: Standard build tool
- **Docker**: Optional, for container image verification
- **golangci-lint**: [golangci-lint.run](https://golangci-lint.run/)

### Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/<your-username>/aws-guardduty-archive-bot.git
   cd aws-guardduty-archive-bot
   ```

2. **Install development dependencies:**
   ```bash
   make dev-setup
   ```

3. **Verify build and tests:**
   ```bash
   make build
   make test
   ```

---

## Development Workflow

### Available Make Targets

- `make build` - Builds the binary output to `bin/`
- `make test` - Runs unit tests
- `make test-coverage` - Runs unit tests with code coverage analysis
- `make fmt` - Formats Go source code using standard Go formatting
- `make lint` - Runs `golangci-lint` to check code quality and conventions
- `make clean` - Cleans temporary build artifacts
- `make docker-build` - Builds the Docker container image

---

## Conventional Commits

This project uses [Release Please](https://github.com/googleapis/release-please) to automate semantic versioning and release notes. All commit messages and Pull Request titles **must** adhere to the [Conventional Commits](https://www.conventionalcommits.org/) specification.

### Format
```
<type>(<optional scope>): <description>
```

### Common Types
- `feat:` A new feature or rule implementation (triggers minor version bump)
- `fix:` A bug fix or patch (triggers patch version bump)
- `docs:` Documentation-only changes
- `refactor:` Code changes that neither fix a bug nor add a feature
- `test:` Adding or updating tests
- `chore:` Tooling, CI/CD, dependency updates, or internal cleanup

---

## Submitting Pull Requests

1. Create a descriptive branch from `main`:
   ```bash
   git checkout -b feat/my-new-feature
   ```
2. Make your changes adhering to standard Go formatting and idioms.
3. Run tests and linter locally:
   ```bash
   make fmt
   make lint
   make test
   ```
4. Commit your changes with a conventional commit message:
   ```bash
   git commit -m "feat(rules): add support for new archive rule"
   ```
5. Push to your fork and open a Pull Request against the `main` branch.
6. Ensure all CI checks pass. Maintainers will review your PR promptly.

---

## Licensing & Contributor Terms

By submitting a pull request, you agree that your contributions will be licensed under the project's [Business Source License 1.1 (BSL 1.1)](LICENSE), with the understanding that they will convert to the Apache License 2.0 under the standard Change Date schedule.
