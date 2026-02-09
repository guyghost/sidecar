package gitstatus

import "github.com/marcus/sidecar/internal/git"

// Re-export commit types from internal/git for backward compatibility.
type CommitError = git.CommitError

// Re-export commit functions.
var (
	ExecuteCommit = git.ExecuteCommit
	ExecuteAmend  = git.ExecuteAmend
)

// GetLastCommitMessage wraps git.GetLastCommitMessage.
func GetLastCommitMessage(workDir string) string {
	return git.GetLastCommitMessage(workDir)
}

// getLastCommitMessage wraps git.GetLastCommitMessage for backward compatibility.
func getLastCommitMessage(workDir string) string {
	return git.GetLastCommitMessage(workDir)
}

// Re-export discard functions.
var (
	DiscardModified  = git.DiscardModified
	DiscardStaged    = git.DiscardStaged
	DiscardUntracked = git.DiscardUntracked
)
