# OpenSustain CLI

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![GitHub Action](https://img.shields.io/badge/GitHub%20Action-OpenSustain-green)](https://github.com/OpenSustain/OpenSustain)

OpenSustain is a Go CLI (and reusable GitHub Action) for generating **maintainer-load and sustainability reports** from local git history and the GitHub API.

---

## Build

```bash
go build -o OpenSustain ./cmd/OpenSustain
```

## CLI Usage

### Scan a single repository

```bash
# Local repo — markdown report to stdout
./OpenSustain scan repo --repo ./myrepo --days 90 --format md

# Local repo — markdown report saved to reports/repo-reports/
./OpenSustain scan repo --repo ./myrepo --days 90 --format md

# Remote GitHub repo — JSON report to file (requires token for issues/PRs)
./OpenSustain scan repo \
  --repo owner/repo \
  --days 90 \
  --format json \
  --out report.json \
  --token "$GITHUB_TOKEN"
```

- **Remote mode (default):** For GitHub repositories, OpenSustain uses only GitHub APIs to compute commits, contributors, backlog, and response time metrics.
- **Deep mode (optional):** Use `--mode deep`, `--local`, or point `--repo` at a local repository path to use git log and file-level history for precise ownership and bus-factor analysis.

### Scan a whole organisation

```bash
./OpenSustain scan org \
  --org my-github-org \
  --days 90 \
  --format md \
  --token "$GITHUB_TOKEN"
```

**Rate Limiting Optimization** (for large organizations):

```bash
# Skip response time entirely (max API savings)
./OpenSustain scan org --org my-github-org --days 90 --skip-response-time --token "$GITHUB_TOKEN"

# Sample 20% of issues for response time
./OpenSustain scan org --org my-github-org --days 90 --sample-rate 0.2 --token "$GITHUB_TOKEN"

# Only fetch response times for recent issues
./OpenSustain scan org --org my-github-org --days 90 --recent-only --token "$GITHUB_TOKEN"
```

If `--out` is omitted, reports are automatically saved to `reports/org-reports/{org-name}-{timestamp}.{format}`.

> **Note:** The `--token` flag is required for org scanning. The token needs `read:org` and `repo` scopes.

---

## GitHub Action Usage

Add OpenSustain to any workflow in `.github/workflows/`.

### Repo-level scan

```yaml
- name: OpenSustain repo scan
  uses: OpenSustain/OpenSustain@main
  with:
    mode: repo
    repo: ${{ github.repository }}
    days: "90"
    format: md
    out: opensustain-report.md
    token: ${{ secrets.GITHUB_TOKEN }}
```

### Org-level scan

```yaml
- name: OpenSustain org scan
  uses: OpenSustain/OpenSustain@main
  with:
    mode: org
    org: my-github-org
    days: "90"
    format: md
    out: opensustain-org-report.md
    token: ${{ secrets.ORG_SCAN_TOKEN }}   # needs read:org scope
```

### Inputs

| Input   | Required | Default | Description |
|---------|----------|---------|-------------|
| `mode`  | Yes | `repo` | `repo` or `org` |
| `repo`  | if mode=repo | — | `owner/repo` identifier |
| `org`   | if mode=org | — | GitHub organisation name |
| `days`  | No | `90` | Activity window in days |
| `format`| No | `md` | `md` or `json` |
| `out`   | No | reports/ folder | Output file path (defaults to reports/repo-reports/ or reports/org-repos/ with timestamp) |
| `token` | Yes | — | GitHub token |

### Outputs

| Output | Description |
|--------|-------------|
| `report-path` | Path of the generated report file |

---

## Sustainability Score

Each repository is scored 0–100 across four dimensions:

| Dimension | Max pts | What is measured |
|-----------|---------|-----------------|
| Bus-factor risk | 30 | Top contributor share of commits |
| Backlog age | 30 | Proportion of issues older than 30 days |
| Commit activity | 20 | Total commits in the scan window |
| Response time | 20 | Median time to first comment on issues/PRs |

**Labels:** `Healthy` (≥70) · `Moderate` (40–69) · `At Risk` (<40)

Repositories scoring `At Risk` are flagged in the `HighRiskRepositories` section of org reports.

---

## Features

- Total commits and unique contributors over a configurable time window
- Top-contributor bus-factor share
- Open Issues and PR backlog sizing
- Backlog age buckets (0-7, 8-30, 31-90, 90+ days)
- Sustainability scoring engine (0-100)
- High-risk repository detection for org scans
- Markdown and JSON output formats
- Docker container + reusable GitHub Action

## License

MIT
