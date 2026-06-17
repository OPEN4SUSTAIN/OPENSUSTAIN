---
layout: default
title: Documentation
---

# OpenSustain Documentation

OpenSustain is a Go-based CLI tool designed to generate maintainer-load reports. It works by analyzing Git history (commits, contributors) and GitHub repository activity (issues, pull requests, backlog sizes, and response times) to provide a clear picture of project health.

## Installation & Build

OpenSustain is written in Go and requires Go 1.22+.

To build the executable, run the following command in the root of the project:

```bash
go build -o OpenSustain ./cmd/OpenSustain
```

This generates a standalone binary named `OpenSustain`.

## Commands and Usage

The CLI supports scanning a repository or a full GitHub organization to generate maintainer load and sustainability reports.

### Scan a Repository

Scans a repository to generate a maintainer load report in either JSON or Markdown format.

```bash
./OpenSustain scan repo --repo <path-or-repo> [flags]
```

**Available Flags:**
- `--repo` **(Required)**: Path to a local repository or GitHub owner/repo identifier
- `--days`: Number of days to scan for activity (default: 90)
- `--format`: Output format: 'md' or 'json' (default: md)
- `--out`: Output file path (default: reports/repo-reports/)
- `--local`: Enable deep analysis using local git history
- `--mode`: Scan mode: 'remote' or 'deep'
- `--token`: GitHub Personal Access Token for API access
- `--skip-response-time`: Skip response time fetching to reduce API calls
- `--sample-rate`: Sample rate for response time fetching (0.0-1.0)
- `--recent-only`: Only fetch response times for recent issues

### Scan an Organization

Scans an entire GitHub organization and aggregates metrics across active repositories.

```bash
./OpenSustain scan org --org <organization> --days 90 --format md --token "$GITHUB_TOKEN"
```

**Available Flags:**
- `--org` **(Required)**: GitHub organization name
- `--days`: Number of days to scan for repo activity (default: 90)
- `--format`: Output format: 'md' or 'json' (default: md)
- `--out`: Output file path (default: reports/org-reports/)
- `--token`: GitHub token for API access
- `--app-id`: GitHub App ID for App authentication (higher rate limits)
- `--private-key-path`: Path to GitHub App private key PEM file
- `--skip-response-time`: Skip response time fetching to reduce API calls
- `--sample-rate`: Sample rate for response time fetching (0.0-1.0)
- `--recent-only`: Only fetch response times for recent issues

## Output Metrics

### Activity Metrics
- **Total Commits**: Number of commits in the time window
- **Unique Contributors**: Number of distinct authors
- **Top Contributor Share**: Percentage of commits by the single most active contributor

### GitHub Backlog
- **Open Issues**: Total count of unresolved issues
- **Open Pull Requests**: Total count of unresolved PRs
- **Median Response Time**: Median duration before the first comment

### Backlog Age Buckets
- `0-7 days`: Fresh, needs attention
- `8-30 days`: Normal backlog
- `31-90 days`: Aging, needs triage
- `90+ days`: Stale, critical risk indicator

### Sustainability Scoring
- **Score range:** 0–100
- **Healthy:** 70 and above
- **Moderate:** 40–69
- **At Risk:** below 40

The score combines bus-factor risk, backlog age, commit activity, and response time into a single view of project sustainability.
