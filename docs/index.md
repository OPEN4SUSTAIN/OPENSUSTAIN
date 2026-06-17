---
layout: home
title: OpenSustain
---

## Features

<ul class="feature-list">
  <li><strong>Bus Factor Analysis</strong> – Top contributor share scoring</li>
  <li><strong>Backlog Aging</strong> – 4 age buckets (0–7, 8–30, 31–90, 90+ days)</li>
  <li><strong>Response Time Tracking</strong> – Median time to first comment</li>
  <li><strong>Unified Sustainability Score</strong> – 0–100 single score</li>
  <li><strong>High-Risk Auto-Detection</strong> – Automatic flagging</li>
  <li><strong>Org-Level Aggregation</strong> – Scan entire orgs in one command</li>
  <li><strong>Local Git Analysis</strong> – Deep mode with git log</li>
  <li><strong>GitHub API Integration</strong> – Issues, PRs, response times</li>
</ul>

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

See [GitHub App Setup]({{ '/github-app-setup' | relative_url }}) for detailed instructions.

## Documentation

- [Full Documentation]({{ '/documentation' | relative_url }})
- [GitHub App Setup Guide]({{ '/github-app-setup' | relative_url }})
- [Rate Limiting Optimization]({{ '/rate-limiting' | relative_url }})
- [Features Overview]({{ '/features' | relative_url }}) <span class="badge">Coming Soon</span>

## License

MIT
