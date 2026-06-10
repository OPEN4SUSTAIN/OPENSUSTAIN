package githubclient

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	Token      string
	HTTPClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// doRequest executes an HTTP request and automatically handles GitHub API
// rate limiting by sleeping until the reset window if the limit is exhausted.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		// 429 Too Many Requests or 403 with rate-limit exhausted
		if resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0") {

			if err := resp.Body.Close(); err != nil {
				log.Printf("Warning: failed to close response body: %v", err)
			}
			resetAt := parseRateLimitReset(resp.Header.Get("X-RateLimit-Reset"))
			waitDur := time.Until(resetAt) + 2*time.Second // small buffer
			if waitDur < 0 {
				waitDur = 10 * time.Second
			}
			log.Printf("GitHub rate limit hit; sleeping %s until reset (attempt %d/%d)", waitDur.Round(time.Second), attempt+1, maxRetries)
			time.Sleep(waitDur)
			continue
		}

		return resp, nil
	}
	return nil, fmt.Errorf("GitHub API rate limit exhausted after %d retries", maxRetries)
}

// parseRateLimitReset parses the X-RateLimit-Reset Unix timestamp header.
func parseRateLimitReset(header string) time.Time {
	if header == "" {
		return time.Now().Add(60 * time.Second)
	}
	unix, err := strconv.ParseInt(header, 10, 64)
	if err != nil {
		return time.Now().Add(60 * time.Second)
	}
	return time.Unix(unix, 0)
}

type Issue struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	PullRequest *struct {
		URL string `json:"url"`
	} `json:"pull_request,omitempty"`
}

type Comment struct {
	CreatedAt time.Time `json:"created_at"`
}

type GitHubStats struct {
	CommitCount          int
	UniqueContributors   int
	TopContributorShare  float64
	OpenIssuesCount      int
	OpenPRsCount         int
	IssueResponseTimes   []time.Duration
	PRResponseTimes      []time.Duration
	IssueCreationDates   []time.Time
}

type Repo struct {
	Name     string    `json:"name"`
	FullName string    `json:"full_name"`
	PushedAt time.Time `json:"pushed_at"`
}

// FetchStats queries the GitHub API for commits, issues, and PRs in the given owner/repo.
func (c *Client) FetchStats(ownerRepo string, days int) (*GitHubStats, error) {
	if c.Token == "" {
		// Gracefully skip if no token is provided
		return nil, nil
	}

	stats := &GitHubStats{}

	commitStats, err := c.fetchCommitSummary(ownerRepo, days)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits: %w", err)
	}
	stats.CommitCount = commitStats.CommitCount
	stats.UniqueContributors = commitStats.UniqueContributors
	stats.TopContributorShare = commitStats.TopContributorShare

	issueStats, err := c.fetchIssueStats(ownerRepo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}
	stats.OpenIssuesCount = issueStats.OpenIssuesCount
	stats.OpenPRsCount = issueStats.OpenPRsCount
	stats.IssueCreationDates = issueStats.IssueCreationDates
	stats.IssueResponseTimes = issueStats.IssueResponseTimes
	stats.PRResponseTimes = issueStats.PRResponseTimes

	return stats, nil
}

// parseIssuesResponse is separated to allow unit testing without HTTP calls
func parseIssuesResponse(issues []Issue) (*GitHubStats, error) {
	stats := &GitHubStats{}

	for _, issue := range issues {
		if issue.PullRequest != nil {
			stats.OpenPRsCount++
		} else {
			stats.OpenIssuesCount++
		}
		stats.IssueCreationDates = append(stats.IssueCreationDates, issue.CreatedAt)
	}

	return stats, nil
}

type commitResponse struct {
	Commit struct {
		Author struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
	} `json:"commit"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author"`
}

type commitSummary struct {
	CommitCount         int
	UniqueContributors  int
	TopContributorShare float64
}

func (c *Client) fetchCommitSummary(ownerRepo string, days int) (*commitSummary, error) {
	since := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)
	authorCounts := make(map[string]int)
	commitCount := 0
	page := 1

	for {
		url := fmt.Sprintf("https://api.github.com/repos/%s/commits?since=%s&per_page=100&page=%d", ownerRepo, since, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("Warning: failed to close response body: %v", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("github api returned status: %d", resp.StatusCode)
		}

		var commits []commitResponse
		if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
			return nil, fmt.Errorf("failed to parse commits: %w", err)
		}

		if len(commits) == 0 {
			break
		}

		for _, commit := range commits {
			commitCount++
			authorKey := "unknown"
			if commit.Author != nil && commit.Author.Login != "" {
				authorKey = commit.Author.Login
			} else if commit.Commit.Author.Email != "" {
				authorKey = commit.Commit.Author.Email
			} else if commit.Commit.Author.Name != "" {
				authorKey = commit.Commit.Author.Name
			}
			authorCounts[authorKey]++
		}

		if len(commits) < 100 {
			break
		}
		page++
	}

	summary := &commitSummary{CommitCount: commitCount}
	if commitCount > 0 {
		maxCommits := 0
		for _, count := range authorCounts {
			if count > maxCommits {
				maxCommits = count
			}
		}
		summary.UniqueContributors = len(authorCounts)
		summary.TopContributorShare = float64(maxCommits) / float64(commitCount)
	}

	return summary, nil
}

type issueStats struct {
	OpenIssuesCount    int
	OpenPRsCount       int
	IssueCreationDates []time.Time
	IssueResponseTimes []time.Duration
	PRResponseTimes    []time.Duration
}

func (c *Client) fetchIssueStats(ownerRepo string) (*issueStats, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=100", ownerRepo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	stats := &issueStats{}
	for _, issue := range issues {
		if issue.PullRequest != nil {
			stats.OpenPRsCount++
			if duration, err := c.fetchFirstResponseTime(ownerRepo, issue.Number, issue.CreatedAt); err == nil && duration != nil {
				stats.PRResponseTimes = append(stats.PRResponseTimes, *duration)
			}
		} else {
			stats.OpenIssuesCount++
			if duration, err := c.fetchFirstResponseTime(ownerRepo, issue.Number, issue.CreatedAt); err == nil && duration != nil {
				stats.IssueResponseTimes = append(stats.IssueResponseTimes, *duration)
			}
		}
		stats.IssueCreationDates = append(stats.IssueCreationDates, issue.CreatedAt)
	}

	return stats, nil
}

// fetchFirstResponseTime makes an extra API call to get the first comment for an issue/PR
func (c *Client) fetchFirstResponseTime(ownerRepo string, issueNumber int, createdAt time.Time) (*time.Duration, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%d/comments?per_page=1", ownerRepo, issueNumber)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var comments []Comment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	if len(comments) > 0 {
		duration := comments[0].CreatedAt.Sub(createdAt)
		return &duration, nil
	}

	return nil, nil
}

// FetchOrgRepos queries the GitHub API for repositories in the given org
func (c *Client) FetchOrgRepos(org string) ([]Repo, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("GitHub token is required to fetch org repos")
	}

	// In a real app we'd want pagination. For simplicity, we just fetch the first page of 100.
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?per_page=100&sort=pushed", org)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repos: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status: %d", resp.StatusCode)
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}

	return repos, nil
}
