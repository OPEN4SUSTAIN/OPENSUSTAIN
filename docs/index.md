---
layout: default
title: OpenSustain
---

# OpenSustain

The most comprehensive open source sustainability and maintainer-load analysis tool.

## Features

- **Bus Factor Analysis** – Top contributor share scoring
- **Backlog Aging** – 4 age buckets (0–7, 8–30, 31–90, 90+ days)
- **Response Time Tracking** – Median time to first comment
- **Unified Sustainability Score** – 0–100 single score
- **High-Risk Auto-Detection** – Automatic flagging
- **Org-Level Aggregation** – Scan entire orgs in one command
- **Local Git Analysis** – Deep mode with git log
- **GitHub API Integration** – Issues, PRs, response times

## Quick Start

### Build

```bash
go build -o OpenSustain ./cmd/OpenSustain
```

### Scan a Repository

```bash
./OpenSustain scan repo --repo owner/repo --days 90 --format md --token "$GITHUB_TOKEN"
```

### Scan an Organization

```bash
./OpenSustain scan org --org my-org --days 90 --format md --token "$GITHUB_TOKEN"
```

## Rate Limiting Optimization

For large organizations, use these flags to reduce GitHub API calls:

```bash
# Skip response time entirely (max API savings)
./OpenSustain scan org --org my-org --days 90 --skip-response-time --token "$GITHUB_TOKEN"

# Sample 20% of issues for response time
./OpenSustain scan org --org my-org --days 90 --sample-rate 0.2 --token "$GITHUB_TOKEN"

# Only fetch response times for recent issues
./OpenSustain scan org --org my-org --days 90 --recent-only --token "$GITHUB_TOKEN"
```

## GitHub App Authentication

For enterprise organizations with 100+ repositories, use GitHub App authentication for higher rate limits (5,000/hour vs 60/hour for PATs):

```bash
./OpenSustain scan org --org my-org --days 90 --app-id 123456 --private-key-path /path/to/private-key.pem
```

See [GitHub App Setup](./github-app-setup/) for detailed instructions.

## Documentation

- [Full Documentation](./documentation/)
- [GitHub App Setup Guide](./github-app-setup/)
- [Rate Limiting Optimization](./rate-limiting/)
- [Features Overview](./features/)

## License

MIT
