package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"opensustain/internal/metrics"
)

func TestRenderJSON(t *testing.T) {
	r := NewRenderer("json", "", "test-repo")
	report := &metrics.MetricsReport{
		TotalCommits:       100,
		UniqueContributors: 5,
		MedianResponseTime: 5 * time.Hour,
	}

	out, err := r.renderJSON(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON and contains our keys
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if parsed["total_commits"].(float64) != 100 {
		t.Errorf("expected total_commits=100")
	}
}

func TestRenderMarkdown(t *testing.T) {
	r := NewRenderer("md", "", "test-repo")
	report := &metrics.MetricsReport{
		TotalCommits:       100,
		TopContributorShare: 0.75,
	}

	out, err := r.renderMarkdown(report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "**Total Commits:** 100") {
		t.Errorf("markdown missing total commits")
	}
	if !strings.Contains(out, "75.00%") {
		t.Errorf("markdown missing formatted top contributor share")
	}
}
