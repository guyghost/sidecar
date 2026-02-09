# OpenSpec: Config Hot Reload via File Watching

**ID**: `td-b7b55a`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P3
**Type**: Feature
**Effort**: 5 story points

## Problem Statement

Changes to `~/.config/sidecar/config.json` require a full application restart to take effect. This is particularly painful for theme customization, keymap tuning, and UI preference adjustments where rapid iteration is desired.

## Objective

Add fsnotify-based file watching on `config.json` with hot reload for non-structural settings. Structural changes (plugin registration, adapter configuration) still require restart.

## Constraints

- Must not cause flickering or partial state during reload
- Must handle rapid successive saves (debounce)
- Must validate config before applying (reject invalid)
- Must not break mid-render (atomic swap)

## Technical Design

### Hot-Reloadable vs Restart-Required Settings

| Setting | Hot Reload | Restart |
|---------|-----------|---------|
| Theme (built-in) | ✅ | |
| Theme (community) | ✅ | |
| Project theme | ✅ | |
| Keymap bindings | ✅ | |
| UI preferences (clock, toast duration) | ✅ | |
| Plugin enable/disable | | ✅ |
| Feature flags | | ✅ |
| Adapter configuration | | ✅ |

### Config Watcher

```go
// internal/config/watcher.go

type Watcher struct {
    path     string
    fsw      *fsnotify.Watcher
    onChange func(old, new *Config)
    logger   *slog.Logger
}

func NewWatcher(path string, onChange func(old, new *Config)) (*Watcher, error)
func (w *Watcher) Start() error
func (w *Watcher) Stop() error
```

### Integration with App

```go
// In cmd/sidecar/main.go or internal/app/
configWatcher, _ := config.NewWatcher(configPath, func(old, new *Config) {
    // Emit Bubble Tea message for hot-reloadable changes
    if old.UI.Theme != new.UI.Theme {
        program.Send(ThemeChangedMsg{Theme: new.UI.Theme})
    }
    if !reflect.DeepEqual(old.Keymap, new.Keymap) {
        program.Send(KeymapChangedMsg{Keymap: new.Keymap})
    }
    if old.UI != new.UI {
        program.Send(UIPrefsChangedMsg{UI: new.UI})
    }
})
```

### Debounce Strategy

- 200ms debounce on fsnotify events (editors often write temp file then rename)
- Validate new config before applying
- If invalid, log warning and keep current config
- Show toast notification: "Config reloaded" or "Config error: ..."

## Acceptance Criteria

- [ ] Changes to theme in config.json are applied within 500ms without restart
- [ ] Changes to keymap bindings are applied without restart
- [ ] Invalid config changes show error toast and keep current config
- [ ] Rapid saves (< 100ms apart) are debounced to single reload
- [ ] `config.Watcher` cleans up fsnotify on `Stop()`
- [ ] Structural changes (feature flags, plugins) show "restart required" toast
- [ ] No rendering glitches during hot reload

## Dependencies

- `td-396c71` (immutable styles) — for atomic theme swap during hot reload
- `td-31fc1f` (keymap type safety) — beneficial but not required

## Risks

- **Medium**: Editor save strategies (write-to-temp then rename vs in-place) may cause double-fire
- **Mitigation**: 200ms debounce + validate before apply

## Scenarios

### Scenario: Theme change via config edit
```
Given sidecar is running with "Default" theme
When the user edits config.json and changes theme to "Dracula"
Then within 500ms the UI updates to Dracula colors
And a toast shows "Config reloaded: theme → Dracula"
```

### Scenario: Invalid config change
```
Given sidecar is running
When the user saves invalid JSON in config.json
Then an error toast shows "Config reload failed: invalid JSON at line 15"
And the current config remains active
```

### Scenario: Structural change notification
```
Given sidecar is running
When the user enables a feature flag in config.json
Then a toast shows "Feature flag change requires restart"
And the feature is NOT enabled until restart
```
