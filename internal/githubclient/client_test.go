package githubclient

import (
	"context"
	"encoding/json"
	"testing"
)

func TestParseIssuesResponse(t *testing.T) {
	jsonData := `[
		{
			"number": 1,
			"title": "Fix a bug",
			"state": "open",
			"created_at": "2023-10-01T10:00:00Z"
		},
		{
			"number": 2,
			"title": "Add a feature",
			"state": "open",
			"created_at": "2023-10-02T10:00:00Z",
			"pull_request": {
				"url": "https://api.github.com/repos/owner/repo/pulls/2"
			}
		},
		{
			"number": 3,
			"title": "Update README",
			"state": "open",
			"created_at": "2023-10-03T10:00:00Z",
			"pull_request": {
				"url": "https://api.github.com/repos/owner/repo/pulls/3"
			}
		}
	]`

	var issues []Issue
	err := json.Unmarshal([]byte(jsonData), &issues)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	stats, err := parseIssuesResponse(issues)
	if err != nil {
		t.Fatalf("unexpected error parsing issues: %v", err)
	}

	if stats.OpenIssuesCount != 1 {
		t.Errorf("expected 1 open issue, got %d", stats.OpenIssuesCount)
	}

	if stats.OpenPRsCount != 2 {
		t.Errorf("expected 2 open PRs, got %d", stats.OpenPRsCount)
	}
}

func TestFetchStatsGracefulSkip(t *testing.T) {
	client := NewClient("") // Empty token
	stats, err := client.FetchStats(context.Background(), "owner/repo", 90, false, 1.0, false)

	if err != nil {
		t.Fatalf("expected no error when token is empty, got %v", err)
	}
	if stats != nil {
		t.Fatalf("expected nil stats when token is empty, got %+v", stats)
	}
}
