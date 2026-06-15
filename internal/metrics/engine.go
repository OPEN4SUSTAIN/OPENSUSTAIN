package metrics

import (
	"sort"
	"time"

	"opensustain/internal/githubclient"
	"opensustain/internal/gitinspector"
)

// ScoringWeights defines the weights for each scoring component
type ScoringWeights struct {
	BusFactor      float64 // Weight for bus factor risk (default: 30)
	BacklogAge     float64 // Weight for backlog age (default: 30)
	CommitActivity float64 // Weight for commit activity (default: 20)
	ResponseTime   float64 // Weight for response time (default: 20)
}

// DefaultWeights returns the default scoring weights
func DefaultWeights() ScoringWeights {
	return ScoringWeights{
		BusFactor:      30,
		BacklogAge:     30,
		CommitActivity: 20,
		ResponseTime:   20,
	}
}

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
	SustainabilityScore int               `json:"sustainability_score"`
}

func ComputeMetrics(gitStats *gitinspector.GitStats, ghStats *githubclient.GitHubStats, now time.Time) *MetricsReport {
	return ComputeMetricsWithWeights(gitStats, ghStats, now, DefaultWeights())
}

func ComputeMetricsWithWeights(gitStats *gitinspector.GitStats, ghStats *githubclient.GitHubStats, now time.Time, weights ScoringWeights) *MetricsReport {
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
	
	report.SustainabilityScore = computeSustainabilityScore(report, weights)
	
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

func computeSustainabilityScore(report *MetricsReport, weights ScoringWeights) int {
	totalWeight := weights.BusFactor + weights.BacklogAge + weights.CommitActivity + weights.ResponseTime
	if totalWeight == 0 {
		return 0
	}
	
	// Bus Factor Score: Lower top contributor share = higher score
	busFactorScore := 0.0
	if report.TopContributorShare > 0 {
		busFactorScore = (1.0 - report.TopContributorShare) * 100
	}
	
	// Backlog Age Score: Fewer stale issues = higher score
	backlogAgeScore := 0.0
	totalIssues := report.OpenIssuesCount + report.OpenPRsCount
	if totalIssues > 0 {
		staleRatio := float64(report.BacklogAgeBuckets.OverNinety) / float64(totalIssues)
		backlogAgeScore = (1.0 - staleRatio) * 100
	} else {
		backlogAgeScore = 100 // No issues is good
	}
	
	// Commit Activity Score: More commits = higher score (simplified)
	commitActivityScore := 0.0
	if report.TotalCommits > 0 {
		// Normalize: assume 100 commits in period is "good"
		commitActivityScore = min(100.0, float64(report.TotalCommits)/10.0)
	}
	
	// Response Time Score: Faster response = higher score
	responseTimeScore := 0.0
	if report.MedianResponseTime > 0 {
		// Assume 24 hours is acceptable baseline
		hours := report.MedianResponseTime.Hours()
		if hours <= 24 {
			responseTimeScore = 100
		} else {
			responseTimeScore = max(0, 100-(hours-24)*2)
		}
	} else {
		responseTimeScore = 50 // Neutral if no data
	}
	
	// Calculate weighted score
	weightedScore := (busFactorScore*weights.BusFactor +
		backlogAgeScore*weights.BacklogAge +
		commitActivityScore*weights.CommitActivity +
		responseTimeScore*weights.ResponseTime) / totalWeight
	
	return int(weightedScore)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
