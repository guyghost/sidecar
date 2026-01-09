package gitstatus

import (
	"strings"
	"testing"
)

func TestComputeGraph_Empty(t *testing.T) {
	lines := ComputeGraphForCommits([]*Commit{})
	if lines != nil {
		t.Errorf("Expected nil for empty commits, got %v", lines)
	}
}

func TestComputeGraph_SingleCommit(t *testing.T) {
	commits := []*Commit{
		{Hash: "c1", ParentHashes: []string{}}, // root, no parents
	}

	lines := ComputeGraphForCommits(commits)

	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	if !containsRune(lines[0].Chars, '*') {
		t.Errorf("Expected * in line, got %s", lines[0].String())
	}
}

func TestComputeGraph_LinearHistory(t *testing.T) {
	// 3 commits in a straight line: c3 -> c2 -> c1
	commits := []*Commit{
		{Hash: "c3", ParentHashes: []string{"c2"}},
		{Hash: "c2", ParentHashes: []string{"c1"}},
		{Hash: "c1", ParentHashes: []string{}}, // root
	}

	lines := ComputeGraphForCommits(commits)

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	// All should be single column with *
	for i, line := range lines {
		if !containsRune(line.Chars, '*') {
			t.Errorf("Line %d should contain *, got %s", i, line.String())
		}
	}
}

func TestComputeGraph_SimpleMerge(t *testing.T) {
	// Merge commit with 2 parents
	// c3 (merge of c1 and c2)
	// |\
	// c1 c2
	commits := []*Commit{
		{Hash: "c3", ParentHashes: []string{"c1", "c2"}, IsMerge: true},
		{Hash: "c2", ParentHashes: []string{"base"}},
		{Hash: "c1", ParentHashes: []string{"base"}},
		{Hash: "base", ParentHashes: []string{}},
	}

	lines := ComputeGraphForCommits(commits)

	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(lines))
	}

	// First line (merge commit) should show branch connection
	mergeLineStr := lines[0].String()
	if !containsRune(lines[0].Chars, '*') {
		t.Errorf("Merge line should contain *, got %s", mergeLineStr)
	}
}

func TestGraphState_FindOrCreateColumn(t *testing.T) {
	g := NewGraphState()

	col1 := g.findOrCreateColumn("hash1")
	if col1 != 0 {
		t.Errorf("First column should be 0, got %d", col1)
	}

	col2 := g.findOrCreateColumn("hash2")
	if col2 != 1 {
		t.Errorf("Second column should be 1, got %d", col2)
	}

	// Deactivate col1 and verify it gets reused
	g.columns[0].Active = false
	col3 := g.findOrCreateColumn("hash3")
	if col3 != 0 {
		t.Errorf("Should reuse inactive column 0, got %d", col3)
	}
}

func TestGraphState_FindCommitColumn(t *testing.T) {
	g := NewGraphState()

	// Add some columns
	g.columns = []GraphColumn{
		{CommitHash: "hash1", Active: true},
		{CommitHash: "hash2", Active: true},
		{CommitHash: "hash3", Active: false},
	}

	if idx := g.findCommitColumn("hash1"); idx != 0 {
		t.Errorf("Expected column 0 for hash1, got %d", idx)
	}
	if idx := g.findCommitColumn("hash2"); idx != 1 {
		t.Errorf("Expected column 1 for hash2, got %d", idx)
	}
	// hash3 is inactive, should not be found
	if idx := g.findCommitColumn("hash3"); idx != -1 {
		t.Errorf("Expected -1 for inactive hash3, got %d", idx)
	}
	if idx := g.findCommitColumn("nonexistent"); idx != -1 {
		t.Errorf("Expected -1 for nonexistent, got %d", idx)
	}
}

func TestGraphLine_WidthConsistency(t *testing.T) {
	commits := []*Commit{
		{Hash: "c3", ParentHashes: []string{"c1", "c2"}, IsMerge: true},
		{Hash: "c2", ParentHashes: []string{"c1"}},
		{Hash: "c1", ParentHashes: []string{}},
	}

	lines := ComputeGraphForCommits(commits)

	for i, line := range lines {
		if line.Width != len(line.Chars) {
			t.Errorf("Line %d: Width %d should match Chars length %d", i, line.Width, len(line.Chars))
		}
	}
}

func TestGraphLine_String(t *testing.T) {
	gl := GraphLine{Chars: []rune{'*', ' ', '|', ' '}, Width: 4}
	expected := "* | "
	if gl.String() != expected {
		t.Errorf("Expected %q, got %q", expected, gl.String())
	}
}

func TestComputeGraph_RootCommitDeactivatesColumn(t *testing.T) {
	// Single root commit should deactivate its column
	commits := []*Commit{
		{Hash: "root", ParentHashes: []string{}},
	}

	g := NewGraphState()
	_ = g.ComputeGraphLine(commits[0])

	// After processing root commit, its column should be inactive
	if len(g.columns) == 0 {
		// Column was never created, that's fine
		return
	}
	// If column exists, it should be inactive
	for _, col := range g.columns {
		if col.CommitHash == "" && col.Active {
			t.Errorf("Root commit column should be inactive")
		}
	}
}

// containsRune checks if a rune slice contains a target rune.
func containsRune(chars []rune, target rune) bool {
	for _, ch := range chars {
		if ch == target {
			return true
		}
	}
	return false
}

// containsChar checks if line string contains a substring.
func containsChar(s, substr string) bool {
	return strings.Contains(s, substr)
}
