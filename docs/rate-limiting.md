---
layout: default
title: Rate Limiting Optimization
---

# Rate Limiting Optimization

OpenSustain provides several strategies to reduce GitHub API calls when scanning large organizations.

## The Problem

GitHub API rate limits can be a bottleneck when scanning organizations with many repositories:
- Each repo requires multiple API calls (commits, issues, PRs, comments)
- Response time fetching makes 1 extra API call per issue/PR
- Large backlogs (1000+ issues) hit rate limits immediately
- PAT rate limit: 60 requests/hour
- GitHub App rate limit: 5,000 requests/hour

## Solution: Configurable Rate Limiting Strategies

### 1. Skip Response Time

Skip response time fetching entirely to eliminate 1 API call per issue/PR.

```bash
./OpenSustain scan org --org my-org --days 90 --skip-response-time --token "$GITHUB_TOKEN"
```

**Impact:** 1000 issues = 1000 fewer API calls

**Use Case:** Maximum API savings when response time is not critical

### 2. Sample Rate

Sample a percentage of issues for response time fetching.

```bash
./OpenSustain scan org --org my-org --days 90 --sample-rate 0.2 --token "$GITHUB_TOKEN"
```

**Impact:** 20% sample rate = 80% fewer API calls

**Use Case:** Statistical approximation of response times

### 3. Recent Only

Only fetch response times for issues within the scan window.

```bash
./OpenSustain scan org --org my-org --days 90 --recent-only --token "$GITHUB_TOKEN"
```

**Impact:** Eliminates API calls for stale issues

**Use Case:** Focus on current activity, ignore historical backlog

### 4. GitHub App Authentication

Use GitHub App for significantly higher rate limits.

```bash
./OpenSustain scan org --org my-org --days 90 --app-id 123456 --private-key-path /path/to/key.pem
```

**Impact:** 83x higher rate limits (5,000/hour vs 60/hour)

**Use Case:** Enterprise organizations with 100+ repositories

## Combined Strategies

You can combine multiple strategies for maximum efficiency:

```bash
./OpenSustain scan org \
  --org my-org \
  --days 90 \
  --app-id 123456 \
  --private-key-path /path/to/key.pem \
  --sample-rate 0.3 \
  --recent-only
```

## Rate Limit Comparison

| Method | Rate Limit | Best For |
|--------|-----------|----------|
| PAT | 60/hour | Small organizations, testing |
| GitHub App | 5,000/hour | Large organizations, enterprise |
| Skip Response Time | N/A | Maximum API savings |
| 20% Sampling | N/A | Statistical approximation |
| Recent Only | N/A | Current activity focus |

## Recommendations

**Small Organizations (< 50 repos):**
- Use PAT authentication
- No rate limiting optimization needed

**Medium Organizations (50-100 repos):**
- Use PAT authentication
- Add `--sample-rate 0.5` for moderate savings

**Large Organizations (100+ repos):**
- Use GitHub App authentication
- Add `--sample-rate 0.3` and `--recent-only` for optimal performance
