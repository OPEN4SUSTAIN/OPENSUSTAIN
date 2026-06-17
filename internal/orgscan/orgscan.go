package orgscan

import (
	"context"
	"fmt"
	"log"
	"time"

	"opensustain/internal/githubclient"
	"opensustain/internal/metrics"
)

// RepoResult holds the scan result and sustainability score for a single repository.
type RepoResult struct {
	RepoName string
	Report   *metrics.MetricsReport
	Score    SustainabilityScore
}

// OrgMetricsReport aggregates per-repo results and an org-level score.
type OrgMetricsReport struct {
	OrgName              string
	ScannedAt            time.Time
	TotalRepos           int
	ActiveRepos          int
	Repos                []RepoResult
	OrgScore             SustainabilityScore
	HighRiskRepositories []string
}

// ScanOrg enumerates the org's repositories, filters by recent activity,
// runs existing repo-scan logic on each, computes sustainability scores,
// detects high-risk repos, and returns an OrgMetricsReport.
func ScanOrg(org string, days int, token string, appID int64, privateKeyPath string, weights metrics.ScoringWeights, skipResponseTime bool, sampleRate float64, recentOnly bool) (*OrgMetricsReport, error) {
	var client *githubclient.Client
	var err error

	// Use GitHub App authentication if provided
	if appID > 0 && privateKeyPath != "" {
		log.Printf("Using GitHub App authentication (App ID: %d)", appID)
		appAuth, err := githubclient.NewAppAuth(appID, privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize GitHub App auth: %w", err)
		}

		// Get installation for the organization
		installation, err := appAuth.GetInstallationByOrg(org)
		if err != nil {
			return nil, fmt.Errorf("failed to get installation for org %s: %w", org, err)
		}

		// Exchange JWT for installation token
		tokenResp, err := appAuth.GetInstallationToken(installation.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get installation token: %w", err)
		}

		log.Printf("Successfully obtained installation token (expires at: %s)", tokenResp.ExpiresAt.Format("2006-01-02 15:04:05"))
		client = githubclient.NewClient(tokenResp.Token)
	} else if token != "" {
		log.Printf("Using PAT authentication")
		client = githubclient.NewClient(token)
	} else {
		return nil, fmt.Errorf("--token or --app-id/--private-key-path is required for org scanning")
	}

	ctx := context.Background()

	log.Printf("Fetching repositories for org: %s", org)
	repos, err := client.FetchOrgRepos(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch org repos: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -days)
	orgReport := &OrgMetricsReport{
		OrgName:    org,
		ScannedAt:  time.Now(),
		TotalRepos: len(repos),
	}

	totalRepos := len(repos)
	for i, repo := range repos {
		// Phase 2: filter repos active within the window
		if repo.PushedAt.Before(cutoff) {
			log.Printf("Skipping inactive repo: %s (last push: %s)", repo.FullName, repo.PushedAt.Format("2006-01-02"))
			continue
		}

		orgReport.ActiveRepos++
		log.Printf("Scanning repo [%d/%d]: %s", i+1, totalRepos, repo.FullName)

		// Remote org scan uses GitHub API only.
		ghStats, ghErr := client.FetchStats(ctx, repo.FullName, days, skipResponseTime, sampleRate, recentOnly)
		if ghErr != nil {
			log.Printf("Warning: GitHub API failed for %s: %v", repo.FullName, ghErr)
		}

		repoMetrics := metrics.ComputeMetricsWithWeights(nil, ghStats, time.Now(), weights)

		result := RepoResult{
			RepoName: repo.FullName,
			Report:   repoMetrics,
		}

		// Phase 3: compute sustainability score
		result.Score = ComputeSustainabilityScore(&result)

		// Phase 4: flag high-risk repos
		if isHighRisk(result.Score) {
			orgReport.HighRiskRepositories = append(orgReport.HighRiskRepositories, repo.FullName)
			log.Printf("High-risk repo detected: %s (score: %d)", repo.FullName, result.Score.Score)
		}

		orgReport.Repos = append(orgReport.Repos, result)
	}

	// Phase 3: compute org-level score (average of active repos)
	orgReport.OrgScore = computeOrgScore(orgReport.Repos)

	return orgReport, nil
}

// computeOrgScore returns an average SustainabilityScore across all repo results.
func computeOrgScore(repos []RepoResult) SustainabilityScore {
	if len(repos) == 0 {
		return SustainabilityScore{Score: 0, Label: "No Data"}
	}
	total := 0
	for _, r := range repos {
		total += r.Score.Score
	}
	avg := total / len(repos)
	return SustainabilityScore{
		Score: avg,
		Label: scoreLabel(avg),
	}
}
