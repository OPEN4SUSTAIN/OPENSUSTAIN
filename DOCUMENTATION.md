# OpenSustain Documentation

OpenSustain is a Go-based CLI tool designed to generate maintainer-load reports. It works by analyzing Git history (commits, contributors) and GitHub repository activity (issues, pull requests, backlog sizes, and response times) to provide a clear picture of project health.

## Installation & Build

OpenSustain is written in Go and requires Go 1.22+.

To build the executable, run the following command in the root of the project:

```bash
go build -o OpenSustain ./cmd/OpenSustain
```

This generates a standalone binary named `OpenSustain`.

---

## Commands and Usage

The CLI supports scanning a repository or a full GitHub organization to generate maintainer load and sustainability reports.

### `scan repo`

Scans a repository to generate a maintainer load report in either JSON or Markdown format.

**Basic Usage:**
```bash
./OpenSustain scan repo --repo <path-or-repo> [flags]
```

**Available Flags:**
- `--repo` **(Required)**: Path to a local repository (e.g., `./`) or a remote GitHub owner/repo identifier (e.g., `kubernetes/kubernetes`).
- `--days`: Number of days to scan for activity backwards from today. Defaults to `90`.
- `--format`: The output format of the report. Must be `'md'` (Markdown) or `'json'`. Defaults to `md`.
- `--out`: Output file path. If omitted, the report is automatically saved to `reports/repo-reports/{repo-name}-{timestamp}.{format}`.
- `--local`: Optional flag to enable deep analysis using local git history. Use this when scanning a local repository path.
- `--mode`: Optional scan mode. Use `remote` for GitHub API-only scans or `deep` for local git analysis.
- `--token`: GitHub Personal Access Token (PAT). Used to authenticate with the GitHub API to pull Issue and PR statistics and to run remote GitHub-only scans.
- `--skip-response-time`: Skip response time fetching to reduce GitHub API calls. Useful for large backlogs or rate-limited scenarios.
- `--sample-rate`: Sample rate for response time fetching (0.0-1.0, default 1.0 = all). Reduces API calls by only fetching response times for a percentage of issues.
- `--recent-only`: Only fetch response times for issues within the scan window. Eliminates API calls for stale issues.

**Modes:**
- **Remote mode (default):** For remote owner/repo values, OpenSustain uses GitHub APIs only. No local repository clone is required.
- **Deep mode:** Pass `--local`, `--mode deep`, or provide a local repository path to compute precise ownership and bus-factor metrics from `git log`.

### `scan org`

Scans an entire GitHub organization and aggregates metrics across active repositories into a single sustainability report.

**Basic Usage:**
```bash
./OpenSustain scan org --org <organization> --days 90 --format md --token "$GITHUB_TOKEN"
```

**Available Flags:**
- `--org` **(Required)**: GitHub organization name to scan.
- `--days`: Number of days to scan for repo activity. Defaults to `90`.
- `--format`: The output format of the report. Must be `'md'` or `'json'`. Defaults to `md`.
- `--out`: Output file path. If omitted, the report is automatically saved to `reports/org-reports/{org-name}-{timestamp}.{format}`.
- `--token`: GitHub token for API access. Required for organization scans.
- `--skip-response-time`: Skip response time fetching to reduce GitHub API calls. Useful for large backlogs or rate-limited scenarios.
- `--sample-rate`: Sample rate for response time fetching (0.0-1.0, default 1.0 = all). Reduces API calls by only fetching response times for a percentage of issues.
- `--recent-only`: Only fetch response times for issues within the scan window. Eliminates API calls for stale issues.

---

## Examples

### 1. Local Repository Scan

To scan the local repository and output a markdown report to `local_report.md`:

```bash
./OpenSustain scan repo --repo ./ --days 90 --format md
```

This will save the report to `reports/repo-repos/{timestamp}.md`. If you want to specify a custom path, use `--out`:

```bash
./OpenSustain scan repo --repo ./ --days 90 --format md --out local_report.md
```

*Note: If the current branch has no commits, the CLI logs a warning and proceeds with a clean report.*

### 2. Remote GitHub Repository Scan

To scan a GitHub repository and include issue and PR backlog details:

```bash
./OpenSustain scan repo --repo my-org/my-repo --days 180 --format json --token ghp_yourSecretTokenHere
```

### 3. Organization Sustainability Scan

To scan all active repositories in a GitHub organization and generate an aggregated org report:

```bash
./OpenSustain scan org --org my-github-org --days 90 --format md --token "$GITHUB_TOKEN"
```

To reduce GitHub API calls for large organizations:

```bash
# Skip response time entirely (max API savings)
./OpenSustain scan org --org my-github-org --days 90 --skip-response-time --token "$GITHUB_TOKEN"

# Sample 20% of issues for response time
./OpenSustain scan org --org my-github-org --days 90 --sample-rate 0.2 --token "$GITHUB_TOKEN"

# Only fetch response times for recent issues
./OpenSustain scan org --org my-github-org --days 90 --recent-only --token "$GITHUB_TOKEN"
```

If `--out` is omitted, the report is automatically saved to `reports/org-reports/{org-name}-{timestamp}.{format}`.

---

## Output Metrics Breakdown

Whether you use Markdown or JSON, the reports track the following key fields:

### Activity Metrics
*Powered by Local Git History*
- **Total Commits**: Number of commits in the given time window.
- **Unique Contributors**: Number of distinct authors in the time window.
- **Top Contributor Share**: The percentage of commits made by the single most active contributor (used to calculate bus-factor risk).

### GitHub Backlog
*Powered by GitHub REST API (requires `--token`)*
- **Open Issues**: Total count of unresolved issues.
- **Open Pull Requests**: Total count of unresolved PRs.
- **Median Response Time**: Median duration before the first comment is added to an issue or PR.

### Backlog Age Buckets
Open issues are categorized by age into actionable buckets:
- `0-7 days`
- `8-30 days`
- `31-90 days`
- `90+ days` (Stale backlog)

### Sustainability Scoring
Each repository and organization report includes a sustainability score to highlight health and risk.
- **Score range:** 0–100
- **Healthy:** 70 and above
- **Moderate:** 40–69
- **At Risk:** below 40

The score combines bus-factor risk, backlog age, commit activity, and response time into a single view of project sustainability.

---

## GitHub Action Wrapper

OpenSustain can also run as a reusable GitHub Action using the included `action.yml` and `Dockerfile`.

### Action inputs
- `mode`: `repo` or `org`
- `repo`: GitHub owner/repo when using `mode: repo`
- `org`: GitHub organization when using `mode: org`
- `days`: Activity window in days
- `format`: `md` or `json`
- `out`: Output file path
- `token`: GitHub token for API access

### Example workflow snippet

```yaml
- name: Run OpenSustain org scan
  uses: ./
  with:
    mode: org
    org: my-github-org
    days: 90
    format: md
    out: opensustain-org-report.md
    token: ${{ secrets.ORG_SCAN_TOKEN }}
```

The Action builds the CLI in Docker and forwards the selected scan arguments to the executable.
