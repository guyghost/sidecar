package app

import (
	"path/filepath"
	"strings"

	"github.com/marcus/sidecar/internal/git"
)

// Re-export git worktree types and functions from internal/git for backward compatibility.
type WorktreeInfo = git.WorktreeInfo

var (
	GetWorktrees         = git.GetWorktrees
	GetMainWorktreePath  = git.GetMainWorktreePath
	GetAllRelatedPaths   = git.GetAllRelatedPaths
	WorktreeNameForPath  = git.WorktreeNameForPath
	GetRepoName          = git.GetRepoName
	WorktreeExists       = git.WorktreeExists
	CheckCurrentWorktree = git.CheckCurrentWorktree
)

// normalizePath wraps git.normalizePath for internal app package use.
func normalizePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return filepath.Clean(absPath), nil
	}
	return filepath.Clean(resolved), nil
}

// parseWorktreeList is a test helper - see git/worktree.go for implementation.
func parseWorktreeList(output string) []WorktreeInfo {
	var worktrees []WorktreeInfo
	var current WorktreeInfo
	isFirst := true

	lines := strings.Split(output, "\n")
	for _, l := range lines {
		line := strings.TrimSpace(l)
		if line == "" {
			if current.Path != "" {
				current.IsMain = isFirst
				worktrees = append(worktrees, current)
				isFirst = false
			}
			current = WorktreeInfo{}
			continue
		}

		if wtPath, found := strings.CutPrefix(line, "worktree "); found {
			current.Path = filepath.Clean(wtPath)
		} else if branchRef, found := strings.CutPrefix(line, "branch "); found {
			current.Branch = strings.TrimPrefix(branchRef, "refs/heads/")
		}
	}

	// Handle last entry if no trailing newline
	if current.Path != "" {
		current.IsMain = isFirst
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// parseRepoNameFromURL is a test helper - see git/worktree.go for implementation.
func parseRepoNameFromURL(url string) string {
	url = strings.TrimSuffix(url, ".git")

	if idx := strings.LastIndex(url, ":"); idx != -1 && !strings.Contains(url, "://") {
		url = url[idx+1:]
	}

	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[idx+1:]
	}

	return url
}
