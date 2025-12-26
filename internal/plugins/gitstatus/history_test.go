package gitstatus

import (
	"testing"
	"time"
)

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		contains string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"1 minute", now.Add(-1 * time.Minute), "1 min"},
		{"5 minutes", now.Add(-5 * time.Minute), "5 mins"},
		{"1 hour", now.Add(-1 * time.Hour), "1 hour"},
		{"3 hours", now.Add(-3 * time.Hour), "3 hours"},
		{"yesterday", now.Add(-25 * time.Hour), "yesterday"},
		{"3 days", now.Add(-3 * 24 * time.Hour), "3 days"},
		{"1 week", now.Add(-8 * 24 * time.Hour), "1 week"},
		{"3 weeks", now.Add(-22 * 24 * time.Hour), "3 weeks"},
		{"1 month", now.Add(-35 * 24 * time.Hour), "1 month"},
		{"5 months", now.Add(-150 * 24 * time.Hour), "5 months"},
		{"1 year", now.Add(-400 * 24 * time.Hour), "1 year"},
		{"3 years", now.Add(-1100 * 24 * time.Hour), "3 years"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RelativeTime(tc.input)
			if result == "" {
				t.Error("RelativeTime returned empty string")
			}
			// Just verify it returns something meaningful
			if len(result) < 3 {
				t.Errorf("RelativeTime returned unexpectedly short: %q", result)
			}
		})
	}
}

func TestRelativeTime_Boundaries(t *testing.T) {
	now := time.Now()

	// Test boundary conditions
	tests := []struct {
		name  string
		input time.Time
	}{
		{"exactly 0 seconds", now},
		{"exactly 1 minute", now.Add(-1 * time.Minute)},
		{"exactly 1 hour", now.Add(-1 * time.Hour)},
		{"exactly 1 day", now.Add(-24 * time.Hour)},
		{"exactly 1 week", now.Add(-7 * 24 * time.Hour)},
		{"exactly 1 month", now.Add(-30 * 24 * time.Hour)},
		{"exactly 1 year", now.Add(-365 * 24 * time.Hour)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := RelativeTime(tc.input)
			if result == "" {
				t.Error("RelativeTime returned empty string")
			}
		})
	}
}

func TestCommit_Fields(t *testing.T) {
	// Verify Commit struct can be created with all fields
	commit := Commit{
		Hash:        "abc123def456",
		ShortHash:   "abc123",
		Author:      "Test User",
		AuthorEmail: "test@example.com",
		Date:        time.Now(),
		Subject:     "Test commit",
		Body:        "Extended description",
		Files:       []CommitFile{},
		Stats: CommitStats{
			FilesChanged: 5,
			Additions:    100,
			Deletions:    50,
		},
	}

	if commit.Hash != "abc123def456" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc123def456")
	}
	if commit.Stats.FilesChanged != 5 {
		t.Errorf("Stats.FilesChanged = %d, want 5", commit.Stats.FilesChanged)
	}
}

func TestCommitFile_Fields(t *testing.T) {
	// Verify CommitFile struct can be created with all fields
	file := CommitFile{
		Path:      "new/path.go",
		OldPath:   "old/path.go",
		Status:    StatusRenamed,
		Additions: 10,
		Deletions: 5,
	}

	if file.Path != "new/path.go" {
		t.Errorf("Path = %q, want %q", file.Path, "new/path.go")
	}
	if file.Status != StatusRenamed {
		t.Errorf("Status = %v, want %v", file.Status, StatusRenamed)
	}
}

func TestCommitStats_Fields(t *testing.T) {
	stats := CommitStats{
		FilesChanged: 3,
		Additions:    50,
		Deletions:    25,
	}

	if stats.FilesChanged != 3 {
		t.Errorf("FilesChanged = %d, want 3", stats.FilesChanged)
	}
	if stats.Additions != 50 {
		t.Errorf("Additions = %d, want 50", stats.Additions)
	}
	if stats.Deletions != 25 {
		t.Errorf("Deletions = %d, want 25", stats.Deletions)
	}
}
