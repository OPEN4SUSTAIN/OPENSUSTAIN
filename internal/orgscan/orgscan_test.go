package orgscan

import (
	"strings"
	"testing"
	"time"

	"opensustain/internal/metrics"
)

// ── Phase 2: org aggregation ─────────────────────────────────────────────────

func TestComputeOrgScore_Empty(t *testing.T) {
	s := computeOrgScore(nil)
	if s.Score != 0 {
		t.Errorf("expected score 0 for empty repos, got %d", s.Score)
	}
	if s.Label != "No Data" {
		t.Errorf("expected label 'No Data', got %q", s.Label)
	}
}

func TestComputeOrgScore_Average(t *testing.T) {
	repos := []RepoResult{
		{Score: SustainabilityScore{Score: 80, Label: "Healthy"}},
		{Score: SustainabilityScore{Score: 40, Label: "Moderate"}},
		{Score: SustainabilityScore{Score: 20, Label: "At Risk"}},
	}
	s := computeOrgScore(repos)
	// Average of 80+40+20 = 140/3 = 46
	if s.Score != 46 {
		t.Errorf("expected org score 46, got %d", s.Score)
	}
	if s.Label != "Moderate" {
		t.Errorf("expected label 'Moderate', got %q", s.Label)
	}
}

// ── Phase 3: sustainability score engine ─────────────────────────────────────

func makeResult(topShare float64, commits int, old, fresh int, responseHours float64) *RepoResult {
	rep := &metrics.MetricsReport{
		TotalCommits:       commits,
		TopContributorShare: topShare,
		BacklogAgeBuckets: metrics.BacklogAgeBuckets{
			ZeroToSeven:       fresh,
			ThirtyOneToNinety: old,
		},
		MedianResponseTime: time.Duration(responseHours * float64(time.Hour)),
	}
	return &RepoResult{Report: rep}
}

func TestComputeSustainabilityScore_Healthy(t *testing.T) {
	r := makeResult(0.2, 60, 0, 5, 24) // low share, many commits, fresh backlog, fast response
	s := ComputeSustainabilityScore(r)
	if s.Score < 70 {
		t.Errorf("expected healthy score (>=70), got %d", s.Score)
	}
	if s.Label != "Healthy" {
		t.Errorf("expected label 'Healthy', got %q", s.Label)
	}
}

func TestComputeSustainabilityScore_AtRisk(t *testing.T) {
	r := makeResult(0.95, 2, 10, 0, 500) // high share, few commits, stale backlog, slow response
	s := ComputeSustainabilityScore(r)
	if s.Score >= 40 {
		t.Errorf("expected at-risk score (<40), got %d", s.Score)
	}
	if s.Label != "At Risk" {
		t.Errorf("expected label 'At Risk', got %q", s.Label)
	}
}

func TestComputeSustainabilityScore_NoBacklog(t *testing.T) {
	r := makeResult(0.5, 15, 0, 0, 0) // no backlog at all
	s := ComputeSustainabilityScore(r)
	// Should pick up "+30" for no backlog
	if s.Score < 40 {
		t.Errorf("expected at least moderate score with no backlog, got %d", s.Score)
	}
}

// ── Phase 4: high-risk detection ─────────────────────────────────────────────

func TestIsHighRisk_AtRisk(t *testing.T) {
	s := SustainabilityScore{Score: 20, Label: "At Risk"}
	if !isHighRisk(s) {
		t.Error("expected isHighRisk=true for 'At Risk'")
	}
}

func TestIsHighRisk_Healthy(t *testing.T) {
	s := SustainabilityScore{Score: 80, Label: "Healthy"}
	if isHighRisk(s) {
		t.Error("expected isHighRisk=false for 'Healthy'")
	}
}

func TestIsHighRisk_Moderate(t *testing.T) {
	s := SustainabilityScore{Score: 50, Label: "Moderate"}
	if isHighRisk(s) {
		t.Error("expected isHighRisk=false for 'Moderate'")
	}
}

// ── Phase 5: org renderer ────────────────────────────────────────────────────

func TestOrgRenderer_Markdown_ContainsSummary(t *testing.T) {
	report := &OrgMetricsReport{
		OrgName:    "test-org",
		ScannedAt:  time.Now(),
		TotalRepos: 3,
		ActiveRepos: 2,
		OrgScore:   SustainabilityScore{Score: 60, Label: "Moderate"},
		HighRiskRepositories: []string{"test-org/risky-repo"},
		Repos: []RepoResult{
			{
				RepoName: "test-org/repo-a",
				Report:   &metrics.MetricsReport{TotalCommits: 20, UniqueContributors: 3},
				Score:    SustainabilityScore{Score: 75, Label: "Healthy"},
			},
			{
				RepoName: "test-org/risky-repo",
				Report:   &metrics.MetricsReport{TotalCommits: 1},
				Score:    SustainabilityScore{Score: 10, Label: "At Risk"},
			},
		},
	}

	renderer := NewOrgRenderer("md", "")
	md, err := renderer.renderMarkdown(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{
		"test-org",
		"3",             // total repos
		"60 / 100",     // org score
		"⚠️",            // high-risk section
		"risky-repo",
		"repo-a",
		"Healthy",
		"At Risk",
	}
	for _, c := range checks {
		if !strings.Contains(md, c) {
			t.Errorf("markdown output missing expected string %q", c)
		}
	}
}

func TestOrgRenderer_JSON_Valid(t *testing.T) {
	report := &OrgMetricsReport{
		OrgName:     "json-org",
		TotalRepos:  1,
		ActiveRepos: 1,
		OrgScore:    SustainabilityScore{Score: 55, Label: "Moderate"},
		Repos: []RepoResult{
			{
				RepoName: "json-org/repo",
				Report:   &metrics.MetricsReport{TotalCommits: 5},
				Score:    SustainabilityScore{Score: 55, Label: "Moderate"},
			},
		},
	}
	renderer := NewOrgRenderer("json", "")
	out, err := renderer.renderJSON(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, `"json-org"`) {
		t.Errorf("JSON output missing org name")
	}
}
