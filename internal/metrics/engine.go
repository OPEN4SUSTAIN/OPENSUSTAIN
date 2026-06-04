package metrics

import (
	"sort"
	"time"

	"opensustain/internal/githubclient"
	"opensustain/internal/gitinspector"
)

type BacklogAgeBuckets struct {
	ZeroToSeven       int `json:"zero_to_seven_days"`
	EightToThirty     int `json:"eight_to_thirty_days"`
	ThirtyOneToNinety int `json:"thirty_one_to_ninety_days"`
	OverNinety        int `json:"over_ninety_days"`
}

type MetricsReport struct {
	TotalCommits        int               `json:"total_commits"`
	UniqueContributors  int               `json:"unique_contributors"`
	TopContributorShare float64           `json:"top_contributor_share"`
	OpenIssuesCount     int               `json:"open_issues_count"`
	OpenPRsCount        int               `json:"open_prs_count"`
	BacklogAgeBuckets   BacklogAgeBuckets `json:"backlog_age_buckets"`
	MedianResponseTime  time.Duration     `json:"median_response_time_ns"`
}

func ComputeMetrics(gitStats *gitinspector.GitStats, ghStats *githubclient.GitHubStats, now time.Time) *MetricsReport {
	report := &MetricsReport{}
	
	if gitStats != nil {
		report.TotalCommits = gitStats.TotalCommits
		report.UniqueContributors = gitStats.UniqueContributors()
		
		if report.TotalCommits > 0 {
			var maxCommits int
			for _, count := range gitStats.Contributors {
				if count > maxCommits {
					maxCommits = count
				}
			}
			report.TopContributorShare = float64(maxCommits) / float64(report.TotalCommits)
		}
	} else if ghStats != nil {
		report.TotalCommits = ghStats.CommitCount
		report.UniqueContributors = ghStats.UniqueContributors
		report.TopContributorShare = ghStats.TopContributorShare
	}
	
	if ghStats != nil {
		report.OpenIssuesCount = ghStats.OpenIssuesCount
		report.OpenPRsCount = ghStats.OpenPRsCount
		report.MedianResponseTime = computeMedian(append(ghStats.IssueResponseTimes, ghStats.PRResponseTimes...))
		report.BacklogAgeBuckets = computeBacklogBuckets(ghStats.IssueCreationDates, now)
	}
	
	return report
}

func computeMedian(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})
	
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func computeBacklogBuckets(creationDates []time.Time, now time.Time) BacklogAgeBuckets {
	buckets := BacklogAgeBuckets{}
	
	for _, created := range creationDates {
		days := int(now.Sub(created).Hours() / 24)
		if days < 0 {
			days = 0 // handle future dates safely
		}
		
		if days <= 7 {
			buckets.ZeroToSeven++
		} else if days <= 30 {
			buckets.EightToThirty++
		} else if days <= 90 {
			buckets.ThirtyOneToNinety++
		} else {
			buckets.OverNinety++
		}
	}
	
	return buckets
}
