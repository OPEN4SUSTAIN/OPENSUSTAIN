package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"opensustain/internal/metrics"
)

type Renderer struct {
	Format string
	Out    string
}

func NewRenderer(format, out string) *Renderer {
	return &Renderer{
		Format: format,
		Out:    out,
	}
}

func (r *Renderer) Render(report *metrics.MetricsReport) error {
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

func (r *Renderer) renderJSON(report *metrics.MetricsReport) (string, error) {
	bytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *Renderer) renderMarkdown(report *metrics.MetricsReport) (string, error) {
	md := fmt.Sprintf("# OpenSustain Maintainer Load Report\n\n")
	md += fmt.Sprintf("## Activity Metrics\n")
	md += fmt.Sprintf("- **Total Commits:** %d\n", report.TotalCommits)
	md += fmt.Sprintf("- **Unique Contributors:** %d\n", report.UniqueContributors)
	md += fmt.Sprintf("- **Top Contributor Share:** %.2f%%\n\n", report.TopContributorShare*100)
	
	md += fmt.Sprintf("## GitHub Backlog\n")
	md += fmt.Sprintf("- **Open Issues:** %d\n", report.OpenIssuesCount)
	md += fmt.Sprintf("- **Open Pull Requests:** %d\n", report.OpenPRsCount)
	md += fmt.Sprintf("- **Median Response Time:** %s\n\n", report.MedianResponseTime.String())
	
	md += fmt.Sprintf("### Backlog Age Buckets\n")
	md += fmt.Sprintf("| Age | Count |\n")
	md += fmt.Sprintf("|---|---|\n")
	md += fmt.Sprintf("| 0-7 days | %d |\n", report.BacklogAgeBuckets.ZeroToSeven)
	md += fmt.Sprintf("| 8-30 days | %d |\n", report.BacklogAgeBuckets.EightToThirty)
	md += fmt.Sprintf("| 31-90 days | %d |\n", report.BacklogAgeBuckets.ThirtyOneToNinety)
	md += fmt.Sprintf("| 90+ days | %d |\n", report.BacklogAgeBuckets.OverNinety)
	
	return md, nil
}
