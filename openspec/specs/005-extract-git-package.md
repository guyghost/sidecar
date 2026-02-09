# OpenSpec: Extract internal/git Package

**ID**: `td-7e3911`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Refactoring
**Effort**: 8 story points

## Problem Statement

All git operations live inside `internal/plugins/gitstatus/` (48 files), tightly coupled to the plugin's UI state. Other plugins (workspace, filebrowser) need git operations but cannot reuse this code without importing the gitstatus plugin. This creates hidden dependencies and prevents clean separation of concerns.

## Objective

Extract pure git operations into `internal/git/` as a reusable library. The gitstatus plugin becomes a thin UI layer over `internal/git/`.

## Constraints

- No behavioral changes to any plugin
- `internal/git/` must have zero Bubble Tea dependencies
- `internal/git/` must have zero lipgloss/styles dependencies
- All existing gitstatus tests must pass
- Git operations must remain async-safe (called from tea.Cmd)

## Technical Design

### Package Structure

```
internal/git/
├── git.go           // Core types: Status, DiffFile, DiffHunk, Branch, Commit, Stash
├── status.go        // GitStatus() → []FileStatus
├── diff.go          // Diff(path) → DiffResult, DiffStaged(path) → DiffResult
├── diff_parser.go   // ParseUnifiedDiff(string) → []DiffHunk
├── log.go           // Log(opts) → []Commit, LogFile(path) → []Commit
├── branch.go        // Branches() → []Branch, CurrentBranch() → string
├── operations.go    // Stage, Unstage, Commit, Push, Pull, Fetch, Stash
├── graph.go         // CommitGraph rendering (ASCII art)
├── watcher.go       // GitWatcher for .git directory changes
├── remote.go        // Remote info, tracking status
└── git_test.go
```

### Key Types

```go
package git

type Repository struct {
    root string
    // No UI state, no Bubble Tea types
}

func Open(path string) (*Repository, error)

func (r *Repository) Status() ([]FileStatus, error)
func (r *Repository) Diff(path string, staged bool) (*DiffResult, error)
func (r *Repository) Log(opts LogOptions) ([]Commit, error)
func (r *Repository) Stage(paths ...string) error
func (r *Repository) Unstage(paths ...string) error
func (r *Repository) Commit(message string, opts CommitOptions) error
func (r *Repository) Push(opts PushOptions) error
func (r *Repository) Pull(opts PullOptions) error
func (r *Repository) Branches() ([]Branch, error)
func (r *Repository) Stash(action StashAction) error
func (r *Repository) Watch() (*Watcher, error)
```

### Migration Strategy

1. **Copy** git operation functions from gitstatus to `internal/git/`
2. **Remove** UI-specific logic (lipgloss, tea.Cmd wrappers)
3. **Update** gitstatus plugin to call `internal/git/` functions
4. **Update** workspace plugin to use `internal/git/` directly
5. **Delete** duplicated code from gitstatus

### What Stays in gitstatus Plugin

- View rendering (diff renderer, status renderer, history renderer)
- Bubble Tea message types and Cmd wrappers
- UI state management (view modes, selection, scroll)
- Syntax highlighting for diff content

## Acceptance Criteria

- [ ] `internal/git/` package exists with all listed operations
- [ ] `internal/git/` has zero imports from `charmbracelet/*`
- [ ] `internal/git/` has zero imports from `internal/styles`
- [ ] `go test ./internal/git/...` passes with >80% coverage
- [ ] `go test ./internal/plugins/gitstatus/...` passes (no regressions)
- [ ] workspace plugin can import `internal/git/` without importing gitstatus
- [ ] All git operations (status, diff, stage, commit, push, pull, stash) work via `internal/git/`
- [ ] `DiffParser` is reusable from `internal/git/` for any consumer

## Dependencies

- None (but benefits `007-inline-editor-host` and workspace plugin improvements)

## Risks

- **Medium**: Entangled UI state in current git operations
- **Mitigation**: Start with pure operations (status, diff, log), migrate interactive ops later
- **Medium**: Performance — ensure git process spawning patterns are preserved
- **Mitigation**: Keep the same `exec.Command` patterns, just relocate them

## Scenarios

### Scenario: Workspace plugin uses git directly
```
Given the workspace plugin needs to check branch status
When it calls git.Open(worktreePath).CurrentBranch()
Then it gets the branch name without importing gitstatus plugin
```

### Scenario: Diff parsing reused by filebrowser
```
Given the filebrowser wants to show inline git blame
When it calls git.Open(root).Log(LogOptions{Path: "file.go"})
Then it gets commit history for that file
Without any UI dependency
```
