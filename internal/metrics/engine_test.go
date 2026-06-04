package metrics

import (
	"testing"
	"time"

	"opensustain/internal/githubclient"
	"opensustain/internal/gitinspector"
)

func TestComputeMetrics(t *testing.T) {
	// 1. Mock GitStats
	gitStats := gitinspector.NewGitStats()
	gitStats.TotalCommits = 10
	alice := gitinspector.Author{Name: "Alice", Email: "alice@test.com"}
	bob := gitinspector.Author{Name: "Bob", Email: "bob@test.com"}
	gitStats.Contributors[alice] = 8 // Top contributor
	gitStats.Contributors[bob] = 2

	// 2. Mock GitHubStats
	now := time.Now()
	ghStats := &githubclient.GitHubStats{
		OpenIssuesCount: 3,
		OpenPRsCount:    1,
		IssueResponseTimes: []time.Duration{
			1 * time.Hour,
			2 * time.Hour,
			5 * time.Hour,
		},
		IssueCreationDates: []time.Time{
			now.Add(-2 * 24 * time.Hour),   // 2 days ago (0-7 bucket)
			now.Add(-15 * 24 * time.Hour),  // 15 days ago (8-30 bucket)
			now.Add(-40 * 24 * time.Hour),  // 40 days ago (31-90 bucket)
			now.Add(-100 * 24 * time.Hour), // 100 days ago (90+ bucket)
		},
	}

	report := ComputeMetrics(gitStats, ghStats, now)

	// Validate Git Metrics
	if report.TotalCommits != 10 {
		t.Errorf("Expected 10 total commits, got %d", report.TotalCommits)
	}
	if report.UniqueContributors != 2 {
		t.Errorf("Expected 2 unique contributors, got %d", report.UniqueContributors)
	}
	if report.TopContributorShare != 0.8 {
		t.Errorf("Expected top contributor share 0.8, got %f", report.TopContributorShare)
	}

	// Validate GitHub Metrics
	if report.OpenIssuesCount != 3 {
		t.Errorf("Expected 3 open issues, got %d", report.OpenIssuesCount)
	}
	if report.OpenPRsCount != 1 {
		t.Errorf("Expected 1 open PR, got %d", report.OpenPRsCount)
	}

	// Validate Median (1h, 2h, 5h -> median is 2h)
	if report.MedianResponseTime != 2*time.Hour {
		t.Errorf("Expected median 2h, got %v", report.MedianResponseTime)
	}

	// Validate Age Buckets
	if report.BacklogAgeBuckets.ZeroToSeven != 1 {
		t.Errorf("Expected 1 item in 0-7 bucket, got %d", report.BacklogAgeBuckets.ZeroToSeven)
	}
	if report.BacklogAgeBuckets.EightToThirty != 1 {
		t.Errorf("Expected 1 item in 8-30 bucket, got %d", report.BacklogAgeBuckets.EightToThirty)
	}
	if report.BacklogAgeBuckets.ThirtyOneToNinety != 1 {
		t.Errorf("Expected 1 item in 31-90 bucket, got %d", report.BacklogAgeBuckets.ThirtyOneToNinety)
	}
	if report.BacklogAgeBuckets.OverNinety != 1 {
		t.Errorf("Expected 1 item in 90+ bucket, got %d", report.BacklogAgeBuckets.OverNinety)
	}
}

func TestComputeMedian(t *testing.T) {
	even := []time.Duration{1 * time.Hour, 3 * time.Hour, 5 * time.Hour, 7 * time.Hour} // median (3+5)/2 = 4
	odd := []time.Duration{1 * time.Hour, 3 * time.Hour, 5 * time.Hour}                 // median 3
	empty := []time.Duration{}

	if med := computeMedian(even); med != 4*time.Hour {
		t.Errorf("Even median expected 4h, got %v", med)
	}
	if med := computeMedian(odd); med != 3*time.Hour {
		t.Errorf("Odd median expected 3h, got %v", med)
	}
	if med := computeMedian(empty); med != 0 {
		t.Errorf("Empty median expected 0, got %v", med)
	}
}
