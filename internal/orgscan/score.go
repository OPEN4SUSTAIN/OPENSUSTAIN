package orgscan

// SustainabilityScore represents a health/sustainability score for a repo.
// Score is 0–100 (higher is healthier).
type SustainabilityScore struct {
	Score    int
	Label    string // "Healthy", "Moderate", "At Risk"
	Details  string
}


// ComputeSustainabilityScore calculates a 0–100 score from a MetricsReport.
func ComputeSustainabilityScore(r *RepoResult) SustainabilityScore {
	report := r.Report
	score := 0
	details := ""

	// 1. Contributor concentration (bus-factor risk)
	// Lower top-contributor share is better
	switch {
	case report.TopContributorShare <= 0.3:
		score += 30
		details += "Low bus-factor risk (+30). "
	case report.TopContributorShare <= 0.6:
		score += 15
		details += "Moderate bus-factor risk (+15). "
	default:
		score += 0
		details += "High bus-factor risk (+0). "
	}

	// 2. Backlog age — reward fewer old issues
	old := report.BacklogAgeBuckets.ThirtyOneToNinety + report.BacklogAgeBuckets.OverNinety
	total := report.BacklogAgeBuckets.ZeroToSeven + report.BacklogAgeBuckets.EightToThirty + old
	if total == 0 {
		score += 30 // no backlog at all
		details += "No open backlog (+30). "
	} else {
		ratio := float64(old) / float64(total)
		switch {
		case ratio <= 0.2:
			score += 30
			details += "Fresh backlog (+30). "
		case ratio <= 0.5:
			score += 15
			details += "Aging backlog (+15). "
		default:
			score += 0
			details += "Stale backlog (+0). "
		}
	}

	// 3. Activity — reward commits
	switch {
	case report.TotalCommits >= 50:
		score += 20
		details += "Active codebase (+20). "
	case report.TotalCommits >= 10:
		score += 10
		details += "Moderate activity (+10). "
	default:
		score += 0
		details += "Low activity (+0). "
	}

	// 4. Response time — reward quick maintainer response
	responseHours := report.MedianResponseTime.Hours()
	switch {
	case responseHours == 0:
		score += 10 // no data — neutral
		details += "No response data (+10). "
	case responseHours <= 48:
		score += 20
		details += "Fast response time (+20). "
	case responseHours <= 168:
		score += 10
		details += "Moderate response time (+10). "
	default:
		score += 0
		details += "Slow response time (+0). "
	}

	label := scoreLabel(score)
	return SustainabilityScore{Score: score, Label: label, Details: details}
}

func scoreLabel(score int) string {
	switch {
	case score >= 70:
		return "Healthy"
	case score >= 40:
		return "Moderate"
	default:
		return "At Risk"
	}
}

// isHighRisk returns true if a repo is considered high-risk.
func isHighRisk(s SustainabilityScore) bool {
	return s.Label == "At Risk"
}
