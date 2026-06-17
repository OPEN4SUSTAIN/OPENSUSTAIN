package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"opensustain/internal/githubclient"
	"opensustain/internal/gitinspector"
	"opensustain/internal/metrics"
	"opensustain/internal/orgscan"
	"opensustain/internal/report"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	switch os.Args[1] {
	case "scan":
		scanCmd()
	case "version", "--version", "-v":
		fmt.Printf("OpenSustain version %s\n", version)
		os.Exit(0)
	case "help", "--help", "-h":
		printUsageAndExit()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsageAndExit()
	}
}

func printUsageAndExit() {
	fmt.Fprintf(os.Stderr, "OpenSustain - Maintainer-load report CLI\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  OpenSustain <command> [arguments]\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  scan      Scan a repository (subcommands: repo, org)\n")
	fmt.Fprintf(os.Stderr, "  report    Generate a report (placeholder)\n")
	fmt.Fprintf(os.Stderr, "  version   Print version information\n")
	fmt.Fprintf(os.Stderr, "  help      Print this help message\n\n")
	fmt.Fprintf(os.Stderr, "Use \"OpenSustain <command> -h\" for more information about a command.\n")
	os.Exit(1)
}

func scanCmd() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "The 'scan' command requires a subcommand (e.g., repo).\n")
		os.Exit(1)
	}

	switch os.Args[2] {
	case "repo":
		scanRepoCmd()
	case "org":
		scanOrgCmd()
	default:
		fmt.Fprintf(os.Stderr, "Unknown scan subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func scanRepoCmd() {
	scanRepoFlags := flag.NewFlagSet("scan repo", flag.ExitOnError)

	scanRepoFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of 'scan repo':\n")
		fmt.Fprintf(os.Stderr, "  OpenSustain scan repo [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		scanRepoFlags.PrintDefaults()
	}

	repo := scanRepoFlags.String("repo", "", "Path to local repository or GitHub owner/repo identifier (Required)")
	days := scanRepoFlags.Int("days", 90, "Number of days to scan for activity")
	format := scanRepoFlags.String("format", "md", "Output format: 'json' or 'md'")
	out := scanRepoFlags.String("out", "", "Output file path (default is stdout)")
	mode := scanRepoFlags.String("mode", "remote", "Scan mode: 'remote' or 'deep'")
	local := scanRepoFlags.Bool("local", false, "Use local git repository data for deep analysis")
	token := scanRepoFlags.String("token", "", "GitHub token for API access (optional)")
	busFactorWeight := scanRepoFlags.Float64("bus-factor-weight", 30, "Weight for bus factor risk (default: 30)")
	backlogAgeWeight := scanRepoFlags.Float64("backlog-age-weight", 30, "Weight for backlog age (default: 30)")
	commitActivityWeight := scanRepoFlags.Float64("commit-activity-weight", 20, "Weight for commit activity (default: 20)")
	responseTimeWeight := scanRepoFlags.Float64("response-time-weight", 20, "Weight for response time (default: 20)")
	skipResponseTime := scanRepoFlags.Bool("skip-response-time", false, "Skip response time fetching to reduce API calls")
	sampleRate := scanRepoFlags.Float64("sample-rate", 1.0, "Sample rate for response time fetching (0.0-1.0, default 1.0 = all)")
	recentOnly := scanRepoFlags.Bool("recent-only", false, "Only fetch response times for issues within the scan window")

	if err := scanRepoFlags.Parse(os.Args[3:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		scanRepoFlags.Usage()
		os.Exit(1)
	}

	if *repo == "" {
		fmt.Fprintf(os.Stderr, "Error: the --repo flag is required.\n")
		scanRepoFlags.Usage()
		os.Exit(1)
	}

	if *format != "json" && *format != "md" {
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Must be 'json' or 'md'.\n", *format)
		os.Exit(1)
	}

	if !strings.EqualFold(*mode, "remote") && !strings.EqualFold(*mode, "deep") {
		fmt.Fprintf(os.Stderr, "Error: invalid mode '%s'. Must be 'remote' or 'deep'.\n", *mode)
		scanRepoFlags.Usage()
		os.Exit(1)
	}

	if *sampleRate < 0 || *sampleRate > 1 {
		fmt.Fprintf(os.Stderr, "Error: invalid sample-rate '%f'. Must be between 0.0 and 1.0.\n", *sampleRate)
		scanRepoFlags.Usage()
		os.Exit(1)
	}

	log.Printf("Starting scan for repo: %s (window: %d days)", *repo, *days)

	useLocal := *local || strings.EqualFold(*mode, "deep") || isLocalPath(*repo)
	if strings.EqualFold(*mode, "deep") && !useLocal {
		fmt.Fprintf(os.Stderr, "Error: deep mode requires a local repository path or --local.\n")
		scanRepoFlags.Usage()
		os.Exit(1)
	}

	var gitStats *gitinspector.GitStats
	var err error
	if useLocal {
		gitStats, err = gitinspector.AnalyzeRepo(*repo, *days)
		if err != nil {
			log.Fatalf("Error analyzing local git repository: %v", err)
		}
		log.Printf("Successfully analyzed local git history.")
	} else {
		log.Printf("Remote mode enabled: skipping local git analysis.")
	}

	var ghStats *githubclient.GitHubStats
	if *token != "" {
		client := githubclient.NewClient(*token)
		ghStats, err = client.FetchStats(context.Background(), *repo, *days, *skipResponseTime, *sampleRate, *recentOnly)
		if err != nil {
			log.Printf("Warning: GitHub API ingestion failed: %v", err)
		} else {
			log.Printf("Successfully analyzed GitHub API issues and PRs.")
		}
	} else {
		if !useLocal {
			log.Printf("No GitHub token provided, skipping remote GitHub metrics.")
		} else {
			log.Printf("No GitHub token provided, running local-only analysis.")
		}
	}

	// 3. Metrics engine
	weights := metrics.ScoringWeights{
		BusFactor:      *busFactorWeight,
		BacklogAge:     *backlogAgeWeight,
		CommitActivity: *commitActivityWeight,
		ResponseTime:   *responseTimeWeight,
	}
	reportData := metrics.ComputeMetricsWithWeights(gitStats, ghStats, time.Now(), weights)

	// 4. Report rendering
	renderer := report.NewRenderer(*format, *out, *repo)
	if err := renderer.Render(reportData); err != nil {
		log.Fatalf("Error rendering report: %v", err)
	}

	log.Println("Scan completed successfully.")
}

func scanOrgCmd() {
	scanOrgFlags := flag.NewFlagSet("scan org", flag.ExitOnError)

	scanOrgFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of 'scan org':\n")
		fmt.Fprintf(os.Stderr, "  OpenSustain scan org [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		scanOrgFlags.PrintDefaults()
	}

	org := scanOrgFlags.String("org", "", "GitHub organization name (Required)")
	days := scanOrgFlags.Int("days", 90, "Number of days to scan for activity")
	format := scanOrgFlags.String("format", "md", "Output format: 'json' or 'md'")
	out := scanOrgFlags.String("out", "", "Output file path (default is stdout)")
	token := scanOrgFlags.String("token", "", "GitHub token for API access (optional)")
	busFactorWeight := scanOrgFlags.Float64("bus-factor-weight", 30, "Weight for bus factor risk (default: 30)")
	backlogAgeWeight := scanOrgFlags.Float64("backlog-age-weight", 30, "Weight for backlog age (default: 30)")
	commitActivityWeight := scanOrgFlags.Float64("commit-activity-weight", 20, "Weight for commit activity (default: 20)")
	responseTimeWeight := scanOrgFlags.Float64("response-time-weight", 20, "Weight for response time (default: 20)")
	skipResponseTime := scanOrgFlags.Bool("skip-response-time", false, "Skip response time fetching to reduce API calls")
	sampleRate := scanOrgFlags.Float64("sample-rate", 1.0, "Sample rate for response time fetching (0.0-1.0, default 1.0 = all)")
	recentOnly := scanOrgFlags.Bool("recent-only", false, "Only fetch response times for issues within the scan window")

	if err := scanOrgFlags.Parse(os.Args[3:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		scanOrgFlags.Usage()
		os.Exit(1)
	}

	if *org == "" {
		fmt.Fprintf(os.Stderr, "Error: the --org flag is required.\n")
		scanOrgFlags.Usage()
		os.Exit(1)
	}

	if *format != "json" && *format != "md" {
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Must be 'json' or 'md'.\n", *format)
		os.Exit(1)
	}

	if *sampleRate < 0 || *sampleRate > 1 {
		fmt.Fprintf(os.Stderr, "Error: invalid sample-rate '%f'. Must be between 0.0 and 1.0.\n", *sampleRate)
		scanOrgFlags.Usage()
		os.Exit(1)
	}

	log.Printf("Starting org scan for: %s (window: %d days)", *org, *days)

	weights := metrics.ScoringWeights{
		BusFactor:      *busFactorWeight,
		BacklogAge:     *backlogAgeWeight,
		CommitActivity: *commitActivityWeight,
		ResponseTime:   *responseTimeWeight,
	}
	orgReport, err := orgscan.ScanOrg(*org, *days, *token, weights, *skipResponseTime, *sampleRate, *recentOnly)
	if err != nil {
		log.Fatalf("Error scanning org: %v", err)
	}

	renderer := orgscan.NewOrgRenderer(*format, *out, *org)
	if err := renderer.Render(orgReport); err != nil {
		log.Fatalf("Error rendering org report: %v", err)
	}

	log.Printf("Org scan complete. %d active repos scanned, %d high-risk detected.",
		orgReport.ActiveRepos, len(orgReport.HighRiskRepositories))
}

func isLocalPath(path string) bool {
	if path == "" {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}
