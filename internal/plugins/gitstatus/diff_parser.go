package gitstatus

import (
	"github.com/marcus/sidecar/internal/git"
)

// Re-export diff types from internal/git for backward compatibility.
type (
	LineType     = git.LineType
	WordSegment  = git.WordSegment
	DiffLine     = git.DiffLine
	Hunk         = git.Hunk
	ParsedDiff   = git.ParsedDiff
	FileDiffInfo = git.FileDiffInfo
)

// MultiFileDiff wraps git.MultiFileDiff to add UI rendering methods.
type MultiFileDiff struct {
	*git.MultiFileDiff
}

// Re-export diff constants.
const (
	LineContext = git.LineContext
	LineAdd     = git.LineAdd
	LineRemove  = git.LineRemove
)

// Re-export diff functions.
var (
	ParseUnifiedDiff = git.ParseUnifiedDiff
)

// ParseMultiFileDiff wraps git.ParseMultiFileDiff and returns our local MultiFileDiff.
func ParseMultiFileDiff(diff string) *MultiFileDiff {
	return &MultiFileDiff{MultiFileDiff: git.ParseMultiFileDiff(diff)}
}
