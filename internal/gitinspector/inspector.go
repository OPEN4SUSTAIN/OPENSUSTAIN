package gitinspector

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Author struct {
	Name  string
	Email string
}

type GitStats struct {
	TotalCommits int
	Contributors map[Author]int
	FileChanges  map[string]map[Author]int // file path -> author -> commit count
}

func NewGitStats() *GitStats {
	return &GitStats{
		Contributors: make(map[Author]int),
		FileChanges:  make(map[string]map[Author]int),
	}
}

func (s *GitStats) UniqueContributors() int {
	return len(s.Contributors)
}

// AnalyzeRepo runs git log on the given repo path and parses the history.
func AnalyzeRepo(repoPath string, days int) (*GitStats, error) {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	
	cmd := exec.Command("git", "-C", repoPath, "log", 
		fmt.Sprintf("--since=%s", since), 
		"--name-only", 
		"--pretty=format:COMMIT|%H|%an|%ae")
	
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %v, stderr: %s", err, stderr.String())
	}
	
	return parseGitLog(out.String())
}

// parseGitLog parses the custom format of git log into GitStats.
func parseGitLog(logOutput string) (*GitStats, error) {
	stats := NewGitStats()
	scanner := bufio.NewScanner(strings.NewReader(logOutput))
	
	var currentAuthor Author
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		if strings.HasPrefix(line, "COMMIT|") {
			parts := strings.SplitN(line, "|", 4)
			if len(parts) == 4 {
				currentAuthor = Author{
					Name:  strings.TrimSpace(parts[2]),
					Email: strings.TrimSpace(parts[3]),
				}
				stats.TotalCommits++
				stats.Contributors[currentAuthor]++
			}
		} else {
			// It's a file path
			filePath := line
			if _, exists := stats.FileChanges[filePath]; !exists {
				stats.FileChanges[filePath] = make(map[Author]int)
			}
			stats.FileChanges[filePath][currentAuthor]++
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return stats, nil
}
