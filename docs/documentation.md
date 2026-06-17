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

<ul class="feature-list">
  <li><code>--repo</code> <strong>(Required)</strong>: Path to a local repository or GitHub owner/repo identifier</li>
  <li><code>--days</code>: Number of days to scan for activity (default: 90)</li>
  <li><code>--format</code>: Output format: 'md' or 'json' (default: md)</li>
  <li><code>--out</code>: Output file path (default: reports/repo-reports/)</li>
  <li><code>--local</code>: Enable deep analysis using local git history</li>
  <li><code>--mode</code>: Scan mode: 'remote' or 'deep'</li>
  <li><code>--token</code>: GitHub Personal Access Token for API access</li>
  <li><code>--skip-response-time</code>: Skip response time fetching to reduce API calls</li>
  <li><code>--sample-rate</code>: Sample rate for response time fetching (0.0-1.0)</li>
  <li><code>--recent-only</code>: Only fetch response times for recent issues</li>
</ul>

### Scan an Organization

Scans an entire GitHub organization and aggregates metrics across active repositories.

```bash
./OpenSustain scan org --org <organization> --days 90 --format md --token "$GITHUB_TOKEN"
```

**Available Flags:**

<ul class="feature-list">
  <li><code>--org</code> <strong>(Required)</strong>: GitHub organization name</li>
  <li><code>--days</code>: Number of days to scan for repo activity (default: 90)</li>
  <li><code>--format</code>: Output format: 'md' or 'json' (default: md)</li>
  <li><code>--out</code>: Output file path (default: reports/org-reports/)</li>
  <li><code>--token</code>: GitHub token for API access</li>
  <li><code>--app-id</code>: GitHub App ID for App authentication (higher rate limits)</li>
  <li><code>--private-key-path</code>: Path to GitHub App private key PEM file</li>
  <li><code>--skip-response-time</code>: Skip response time fetching to reduce API calls</li>
  <li><code>--sample-rate</code>: Sample rate for response time fetching (0.0-1.0)</li>
  <li><code>--recent-only</code>: Only fetch response times for recent issues</li>
</ul>

## Output Metrics

### Activity Metrics

<ul class="feature-list">
  <li><strong>Total Commits</strong>: Number of commits in the time window</li>
  <li><strong>Unique Contributors</strong>: Number of distinct authors</li>
  <li><strong>Top Contributor Share</strong>: Percentage of commits by the single most active contributor</li>
</ul>

### GitHub Backlog

<ul class="feature-list">
  <li><strong>Open Issues</strong>: Total count of unresolved issues</li>
  <li><strong>Open Pull Requests</strong>: Total count of unresolved PRs</li>
  <li><strong>Median Response Time</strong>: Median duration before the first comment</li>
</ul>

### Backlog Age Buckets

<ul class="feature-list">
  <li><code>0-7 days</code>: Fresh, needs attention</li>
  <li><code>8-30 days</code>: Normal backlog</li>
  <li><code>31-90 days</code>: Aging, needs triage</li>
  <li><code>90+ days</code>: Stale, critical risk indicator</li>
</ul>

### Sustainability Scoring

<ul class="feature-list">
  <li><strong>Score range:</strong> 0–100</li>
  <li><strong>Healthy:</strong> 70 and above</li>
  <li><strong>Moderate:</strong> 40–69</li>
  <li><strong>At Risk:</strong> below 40</li>
</ul>

The score combines bus-factor risk, backlog age, commit activity, and response time into a single view of project sustainability.
