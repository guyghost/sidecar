package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// parseCommitHash extracts the commit hash from git commit output.
// Format: "[branch hash] message"
func parseCommitHash(output string) string {
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		re := regexp.MustCompile(`\[[\w/-]+ ([a-f0-9]+)\]`)
		matches := re.FindStringSubmatch(lines[0])
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// ExecuteCommit executes a git commit with the given message.
// Returns the commit hash on success or an error with git output on failure.
func ExecuteCommit(workDir, message string) (string, error) {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &CommitError{Output: string(output), Err: err}
	}
	return parseCommitHash(string(output)), nil
}

// ExecuteAmend executes a git commit --amend with the given message.
func ExecuteAmend(workDir, message string) (string, error) {
	cmd := exec.Command("git", "commit", "--amend", "-m", message)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", &CommitError{Output: string(output), Err: err}
	}
	return parseCommitHash(string(output)), nil
}

// GetLastCommitMessage returns the message of the most recent commit.
func GetLastCommitMessage(workDir string) string {
	cmd := exec.Command("git", "log", "-1", "--format=%B")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimRight(string(output), "\n")
}

// CommitError wraps a git commit error with its output.
type CommitError struct {
	Output string
	Err    error
}

func (e *CommitError) Error() string {
	return strings.TrimSpace(e.Output)
}

// DiscardModified discards unstaged changes to a modified file.
func DiscardModified(workDir, path string) error {
	cmd := exec.Command("git", "restore", path)
	cmd.Dir = workDir
	return cmd.Run()
}

// DiscardStaged discards staged changes to a file (unstages and restores).
func DiscardStaged(workDir, path string) error {
	// First unstage
	cmd := exec.Command("git", "restore", "--staged", path)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		return err
	}
	// Then restore working tree
	cmd = exec.Command("git", "restore", path)
	cmd.Dir = workDir
	return cmd.Run()
}

// DiscardUntracked removes an untracked file.
func DiscardUntracked(workDir, path string) error {
	fullPath := filepath.Join(workDir, path)
	return os.Remove(fullPath)
}
