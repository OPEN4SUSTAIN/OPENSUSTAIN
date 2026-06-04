package gitinspector

import (
	"testing"
)

func TestParseGitLog(t *testing.T) {
	logOutput := `COMMIT|hash1|Alice|alice@example.com
file1.go
file2.go

COMMIT|hash2|Bob|bob@example.com
file1.go
`
	stats, err := parseGitLog(logOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TotalCommits != 2 {
		t.Errorf("expected 2 total commits, got %d", stats.TotalCommits)
	}
	if stats.UniqueContributors() != 2 {
		t.Errorf("expected 2 unique contributors, got %d", stats.UniqueContributors())
	}

	alice := Author{"Alice", "alice@example.com"}
	bob := Author{"Bob", "bob@example.com"}

	if stats.Contributors[alice] != 1 {
		t.Errorf("expected 1 commit for Alice, got %d", stats.Contributors[alice])
	}

	if stats.FileChanges["file1.go"][alice] != 1 {
		t.Errorf("expected Alice to modify file1.go 1 time, got %d", stats.FileChanges["file1.go"][alice])
	}
	if stats.FileChanges["file1.go"][bob] != 1 {
		t.Errorf("expected Bob to modify file1.go 1 time, got %d", stats.FileChanges["file1.go"][bob])
	}
}
