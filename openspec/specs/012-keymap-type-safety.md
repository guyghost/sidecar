# OpenSpec: Keymap FocusContext Type Safety

**ID**: `td-31fc1f`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Refactoring
**Effort**: 2 story points

## Problem Statement

Focus contexts in the keymap system are raw strings scattered across `internal/keymap/bindings.go` (30+ unique strings like `"global"`, `"git-status"`, `"file-browser-tree"`, `"conversations-sidebar"`) and in `isRootContext()`/`isTextInputContext()` helper functions. Typos in context strings silently fail (bindings don't fire), and there's no way to discover all valid contexts at compile time.

## Objective

Replace raw string contexts with a typed `FocusContext` type and named constants. Add compile-time validation and discoverability.

## Constraints

- All existing keymap bindings must continue working
- User config `keymap` overrides use string context names (must remain compatible)
- Plugin `FocusContext()` method return type may need updating

## Technical Design

### Define Type and Constants

```go
// internal/keymap/context.go

// FocusContext identifies a keymap context for key binding routing.
type FocusContext string

const (
    ContextGlobal             FocusContext = "global"
    ContextProjectSwitcher    FocusContext = "project-switcher"
    ContextWorktreeSwitcher   FocusContext = "worktree-switcher"
    ContextThemeSwitcher      FocusContext = "theme-switcher"
    
    // Git
    ContextGitNoRepo          FocusContext = "git-no-repo"
    ContextGitStatus          FocusContext = "git-status"
    ContextGitDiff            FocusContext = "git-diff"
    ContextGitDiffSideBySide  FocusContext = "git-diff-side-by-side"
    ContextGitCommit          FocusContext = "git-commit"
    ContextGitHistory         FocusContext = "git-history"
    ContextGitHistoryDetail   FocusContext = "git-history-detail"
    ContextGitBranchPicker    FocusContext = "git-branch-picker"
    ContextGitPush            FocusContext = "git-push"
    ContextGitPull            FocusContext = "git-pull"
    ContextGitStash           FocusContext = "git-stash"
    
    // File browser
    ContextFileBrowserTree    FocusContext = "file-browser-tree"
    ContextFileBrowserPreview FocusContext = "file-browser-preview"
    ContextFileBrowserSearch  FocusContext = "file-browser-search"
    ContextFileBrowserTabs    FocusContext = "file-browser-tabs"
    
    // Conversations
    ContextConvSidebar        FocusContext = "conversations-sidebar"
    ContextConvMessages       FocusContext = "conversations-messages"
    ContextConvSearch         FocusContext = "conversations-search"
    
    // Workspace
    ContextWorkspaceSidebar   FocusContext = "workspace-sidebar"
    ContextWorkspaceDetail    FocusContext = "workspace-detail"
    ContextWorkspaceShell     FocusContext = "workspace-shell"
    
    // TD Monitor
    ContextTDMonitor          FocusContext = "td-monitor"
    
    // Notes
    ContextNotesList          FocusContext = "notes-list"
    ContextNotesEditor        FocusContext = "notes-editor"
    ContextNotesSearch        FocusContext = "notes-search"
)
```

### Update Binding Type

```go
type Binding struct {
    Key     string
    Command string
    Context FocusContext // was string
}
```

### Update Plugin Interface

```go
// In internal/plugin/plugin.go
type Plugin interface {
    // ...
    FocusContext() keymap.FocusContext // was string
}
```

### Context Classification

```go
func (fc FocusContext) IsRoot() bool {
    return fc == ContextGlobal // or check a set
}

func (fc FocusContext) IsTextInput() bool {
    switch fc {
    case ContextGitCommit, ContextFileBrowserSearch, 
         ContextConvSearch, ContextNotesSearch:
        return true
    }
    return false
}
```

### User Config Compatibility

User config `keymap` section uses string context names which are parsed into `FocusContext`:

```go
func ParseContext(s string) (FocusContext, error) {
    fc := FocusContext(s)
    if !isKnownContext(fc) {
        return "", fmt.Errorf("unknown keymap context: %q", s)
    }
    return fc, nil
}
```

Unknown contexts in user config produce a warning log (not fatal).

## Acceptance Criteria

- [ ] `FocusContext` type defined with all 30+ named constants
- [ ] `DefaultBindings()` uses `FocusContext` constants (not raw strings)
- [ ] `Plugin.FocusContext()` returns `keymap.FocusContext`
- [ ] `isRootContext()`/`isTextInputContext()` replaced with method receivers
- [ ] User config keymap overrides continue working with string context names
- [ ] Unknown context in user config logs warning (not panic)
- [ ] All plugin `FocusContext()` implementations updated
- [ ] `go vet ./...` and `go build ./...` pass with zero string literal contexts

## Dependencies

- None

## Risks

- **Low**: Wide-reaching but mechanical change
- **Mitigation**: Start with `keymap/` package, then update plugins one by one

## Scenarios

### Scenario: Typo in user keymap config
```
Given a user config with context "git-stauts" (typo)
When sidecar loads the config
Then a warning is logged: "unknown keymap context: git-stauts"
And the binding is skipped (not applied)
And sidecar starts normally
```

### Scenario: Plugin returns typed context
```
Given the git plugin is in diff view mode
When FocusContext() is called
Then it returns keymap.ContextGitDiff (typed constant)
And the keymap registry matches bindings for that context
```
