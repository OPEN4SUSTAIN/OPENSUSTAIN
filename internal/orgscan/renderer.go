package orgscan

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// OrgRenderer renders an OrgMetricsReport in the requested format.
type OrgRenderer struct {
	Format string
	Out    string
}

// NewOrgRenderer creates a new OrgRenderer.
func NewOrgRenderer(format, out string) *OrgRenderer {
	return &OrgRenderer{Format: format, Out: out}
}

// Render writes the org report to the configured output.
func (r *OrgRenderer) Render(report *OrgMetricsReport) error {
	var output string
	var err error

	switch r.Format {
	case "json":
		output, err = r.renderJSON(report)
	case "md":
		output, err = r.renderMarkdown(report)
	default:
		return fmt.Errorf("unsupported format: %s", r.Format)
	}

	if err != nil {
		return err
	}

	var writer io.Writer = os.Stdout
	if r.Out != "" {
		file, err := os.Create(r.Out)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer file.Close()
		writer = file
	}

	_, err = fmt.Fprintln(writer, output)
	return err
}

func (r *OrgRenderer) renderJSON(report *OrgMetricsReport) (string, error) {
	bytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *OrgRenderer) renderMarkdown(report *OrgMetricsReport) (string, error) {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "# OpenSustain Org Report — %s\n\n", report.OrgName)
	fmt.Fprintf(sb, "_Scanned at: %s_\n\n", report.ScannedAt.Format("2006-01-02 15:04 UTC"))

	fmt.Fprintf(sb, "## Summary\n\n")
	fmt.Fprintf(sb, "| Metric | Value |\n|---|---|\n")
	fmt.Fprintf(sb, "| Total Repos | %d |\n", report.TotalRepos)
	fmt.Fprintf(sb, "| Active Repos (within window) | %d |\n", report.ActiveRepos)
	fmt.Fprintf(sb, "| Org Sustainability Score | **%d / 100** (%s) |\n", report.OrgScore.Score, report.OrgScore.Label)
	fmt.Fprintf(sb, "| High-Risk Repos | %d |\n\n", len(report.HighRiskRepositories))

	if len(report.HighRiskRepositories) > 0 {
		fmt.Fprintf(sb, "## ⚠️ High-Risk Repositories\n\n")
		for _, name := range report.HighRiskRepositories {
			fmt.Fprintf(sb, "- `%s`\n", name)
		}
		fmt.Fprintf(sb, "\n")
	}

	fmt.Fprintf(sb, "## Per-Repository Breakdown\n\n")
	fmt.Fprintf(sb, "| Repository | Score | Label | Commits | Contributors | Open Issues | Open PRs |\n")
	fmt.Fprintf(sb, "|---|---|---|---|---|---|---|\n")
	for _, rr := range report.Repos {
		rep := rr.Report
		fmt.Fprintf(sb, "| `%s` | %d | %s | %d | %d | %d | %d |\n",
			rr.RepoName,
			rr.Score.Score,
			rr.Score.Label,
			rep.TotalCommits,
			rep.UniqueContributors,
			rep.OpenIssuesCount,
			rep.OpenPRsCount,
		)
	}
	fmt.Fprintf(sb, "\n")

	return sb.String(), nil
}
