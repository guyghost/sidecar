# OpenSpec: Decompose app/Model God Object

**ID**: `td-b7e1a2`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P1
**Type**: Refactoring
**Effort**: 8 story points

## Problem Statement

`internal/app/model.go` contains a `Model` struct with ~250 fields covering 10+ modals, each following a duplicated pattern (`showX bool`, `xModal *modal.Modal`, `xModalWidth int`, `xMouseHandler *mouse.Handler`). The companion `update.go` is 1700+ lines with a single deeply-nested `Update()` handler. This god object is the primary maintainability bottleneck in the codebase.

## Objective

Decompose `Model` into focused sub-models and `Update()` into domain-specific handlers, reducing the root struct to <80 fields and no single file over 500 lines.

## Constraints

- Zero behavioral change — all existing tests must continue passing
- Plugin interface (`plugin.Plugin`) must remain unchanged
- All 10 modal types must continue working with priority ordering
- Mouse hit testing must remain functional
- Bubble Tea message routing must be preserved

## Technical Design

### Phase 1: Extract ModalManager

Create `internal/app/modal_manager.go`:

```go
// ModalManager encapsulates all app-level modal state and routing.
type ModalManager struct {
    modals map[ModalKind]*ModalState
    active ModalKind
}

// ModalState holds the state for a single modal instance.
type ModalState struct {
    Visible      bool
    Modal        *modal.Modal
    Width        int
    MouseHandler *mouse.Handler
}

func (mm *ModalManager) ActiveModal() ModalKind
func (mm *ModalManager) HasModal() bool
func (mm *ModalManager) Show(kind ModalKind)
func (mm *ModalManager) Hide(kind ModalKind)
func (mm *ModalManager) Get(kind ModalKind) *ModalState
func (mm *ModalManager) HandleKey(msg tea.KeyMsg) (tea.Cmd, bool)
func (mm *ModalManager) HandleMouse(msg tea.MouseMsg) (tea.Cmd, bool)
func (mm *ModalManager) View(width, height int) string
```

This eliminates ~40 fields from Model (4 fields × 10 modals).

### Phase 2: Extract Switcher State

Create `internal/app/switchers.go`:

```go
type ProjectSwitcherState struct { ... }
type WorktreeSwitcherState struct { ... }
type ThemeSwitcherState struct { ... }
```

Each with its own `Update()` and `View()` methods.

### Phase 3: Decompose update.go

Split into domain-specific handlers:

| File | Handler | Responsibility |
|------|---------|---------------|
| `update_modals.go` | `handleModalUpdate()` | All modal key/mouse routing |
| `update_project.go` | `handleProjectSwitch()` | Project/worktree switching |
| `update_theme.go` | `handleThemeSwitch()` | Theme switcher logic |
| `update_plugins.go` | `handlePluginMessages()` | Plugin message forwarding |
| `update_keys.go` | `handleKeyMsg()` | Root key routing (already partially exists) |

### Phase 4: Reduce Model Fields

Target: `Model` retains only:
- Plugin registry + active plugin index
- ModalManager (single field replacing ~40)
- Switcher states (3 fields replacing ~30)
- Core app state (config, keymap, event bus, logger)
- UI state (width, height, UIState)

## Acceptance Criteria

- [ ] `Model` struct has <80 fields (currently ~250)
- [ ] No single `.go` file in `internal/app/` exceeds 500 lines
- [ ] `go test ./internal/app/...` passes with zero changes to test expectations
- [ ] `go test ./...` passes (no regressions in plugins)
- [ ] `ModalManager` handles all 10 modal types with correct priority ordering
- [ ] Mouse hit testing works for all modals and tabs
- [ ] Theme/project/worktree switching works end-to-end
- [ ] Command palette works with modal manager

## Dependencies

- None (pure refactoring, no new features)

## Risks

- **High**: Subtle message routing changes could break modal dismiss behavior
- **Mitigation**: Add integration tests for modal open/close/priority chains before refactoring

## Scenarios

### Scenario: Modal priority ordering preserved
```
Given the help modal is open
When the user presses "?" to open the command palette
Then the palette is shown (higher priority)
And the help modal remains in background
When the user closes the palette
Then the help modal is visible again
```

### Scenario: Project switch with modals open
```
Given the theme switcher modal is open
When a SwitchProjectMsg is received
Then the theme switcher is closed
And the project switch proceeds normally
And all plugins are reinitialized
```

### Scenario: Mouse routing to correct modal
```
Given the project switcher is open
When the user clicks inside the modal bounds
Then the click is routed to the project switcher mouse handler
When the user clicks outside the modal bounds
Then the modal is dismissed
```
