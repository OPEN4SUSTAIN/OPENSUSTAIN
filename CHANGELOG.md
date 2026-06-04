# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of OpenSustain CLI
- Repository scanning with local git history analysis
- Organization-wide scanning across multiple repositories
- GitHub API integration for issue and PR metrics
- Sustainability scoring engine (0-100 scale)
- Bus-factor risk analysis
- Backlog age tracking and categorization
- Response time metrics for issues and PRs
- High-risk repository detection
- Markdown and JSON output formats
- Docker container support
- GitHub Action integration
- Configurable time window for activity analysis
- Graceful handling of missing GitHub tokens
- Comprehensive test coverage

### Security
- No hardcoded credentials or API keys
- Token validation before API calls
- Minimal required permissions for GitHub API access

## [1.0.0] - 2026-06-04

### Added
- Initial public release
- CLI for repository and organization sustainability analysis
- GitHub Action for automated scanning
- Documentation and contribution guidelines
