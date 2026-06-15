package report

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"opensustain/internal/metrics"
)

type Renderer struct {
	Format string
	Out    string
	RepoName string
}

func NewRenderer(format, out, repoName string) *Renderer {
	return &Renderer{
		Format:   format,
		Out:      out,
		RepoName: repoName,
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
	outPath := r.Out
	if outPath == "" && r.RepoName != "" {
		// Generate default path in reports/repo-reports/
		sanitizedName := strings.ReplaceAll(r.RepoName, "/", "-")
		timestamp := time.Now().Format("20060102-150405")
		ext := r.Format
		if ext == "md" {
			ext = "md"
		}
		outPath = filepath.Join("reports", "repo-reports", fmt.Sprintf("%s-%s.%s", sanitizedName, timestamp, ext))
		
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create reports directory: %w", err)
		}
		log.Printf("Writing report to: %s", outPath)
	}
	
	if outPath != "" {
		file, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("failed to open output file: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Warning: failed to close output file: %v", err)
			}
		}()
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
	md := "# OpenSustain Maintainer Load Report\n\n"
	md += "## Activity Metrics\n"
	md += fmt.Sprintf("- **Total Commits:** %d\n", report.TotalCommits)
	md += fmt.Sprintf("- **Unique Contributors:** %d\n", report.UniqueContributors)
	md += fmt.Sprintf("- **Top Contributor Share:** %.2f%%\n\n", report.TopContributorShare*100)
	
	md += "## GitHub Backlog\n"
	md += fmt.Sprintf("- **Open Issues:** %d\n", report.OpenIssuesCount)
	md += fmt.Sprintf("- **Open Pull Requests:** %d\n", report.OpenPRsCount)
	md += fmt.Sprintf("- **Median Response Time:** %s\n\n", report.MedianResponseTime.String())
	
	md += "### Backlog Age Buckets\n"
	md += "| Age | Count |\n"
	md += "|---|---|\n"
	md += fmt.Sprintf("| 0-7 days | %d |\n", report.BacklogAgeBuckets.ZeroToSeven)
	md += fmt.Sprintf("| 8-30 days | %d |\n", report.BacklogAgeBuckets.EightToThirty)
	md += fmt.Sprintf("| 31-90 days | %d |\n", report.BacklogAgeBuckets.ThirtyOneToNinety)
	md += fmt.Sprintf("| 90+ days | %d |\n", report.BacklogAgeBuckets.OverNinety)
	
	return md, nil
}
