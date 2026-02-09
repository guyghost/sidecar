# OpenSpec: Create Reusable InlineEditorHost Component

**ID**: `td-f35907`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P2
**Type**: Refactoring
**Effort**: 5 story points

## Problem Statement

Both `filebrowser` and `notes` plugins embed `tty.Model` for inline text editing with nearly identical wiring code:
- Enter/exit lifecycle with dimension calculation
- Key forwarding to tmux with escape sequence handling
- Exit confirmation dialogs ("Save changes?")
- Mouse forwarding (SGR)
- Click-away detection
- App-level key routing bypass (`ConsumesTextInput()`)
- Adaptive polling with generation counters

This duplication means bug fixes and improvements must be applied in two places.

## Objective

Create a shared `InlineEditorHost` component that encapsulates the tty.Model lifecycle, reducing each plugin's inline editing code to <20 lines of setup.

## Constraints

- Both filebrowser and notes editing must work identically after refactoring
- The component must support both file-based editing (filebrowser) and buffer-based editing (notes)
- Must not introduce any new external dependencies
- Confirmation dialog must be customizable per plugin

## Technical Design

### Component API

```go
// internal/ui/editor/host.go
package editor

type Config struct {
    // Required
    Shell          string   // e.g., "bash", "zsh"
    InitialContent string   // Content to edit
    FilePath       string   // For file-based editing (empty for buffer)
    
    // Optional callbacks
    OnSave    func(content string) tea.Cmd
    OnExit    func(saved bool) tea.Cmd
    OnResize  func(width, height int)
    
    // Customization
    ConfirmExitMessage string // Default: "Save changes before closing?"
    ShowLineNumbers    bool
}

type Host struct {
    config Config
    tty    *tty.Model
    state  hostState // inactive, active, confirming_exit
    
    // Dimensions
    x, y, width, height int
}

func New(cfg Config) *Host
func (h *Host) Enter(x, y, width, height int) tea.Cmd
func (h *Host) Exit() tea.Cmd
func (h *Host) IsActive() bool
func (h *Host) Update(msg tea.Msg) (*Host, tea.Cmd)
func (h *Host) View() string
func (h *Host) ConsumesTextInput() bool
func (h *Host) HandleMouse(msg tea.MouseMsg) tea.Cmd
func (h *Host) Resize(width, height int)
```

### Plugin Integration (filebrowser example)

```go
// Before: ~150 lines of tty wiring code
// After:

func (p *Plugin) enterEditMode(path string) tea.Cmd {
    content, _ := os.ReadFile(path)
    p.editor = editor.New(editor.Config{
        FilePath:  path,
        InitialContent: string(content),
        OnSave: func(content string) tea.Cmd {
            return p.saveFile(path, content)
        },
        OnExit: func(saved bool) tea.Cmd {
            return p.exitEditMode()
        },
    })
    return p.editor.Enter(p.previewX, p.previewY, p.previewWidth, p.previewHeight)
}
```

### Shared Exit Confirmation

The `Host` internally manages the "save changes?" modal using `internal/modal`:

```go
func (h *Host) confirmExit() tea.Cmd {
    h.state = confirmingExit
    h.confirmModal = modal.New(h.config.ConfirmExitMessage).
        AddSection(modal.Buttons(
            modal.Button{Label: "Save", Action: h.saveAndExit},
            modal.Button{Label: "Discard", Action: h.discardAndExit},
            modal.Button{Label: "Cancel", Action: h.cancelExit},
        ))
    return nil
}
```

## Acceptance Criteria

- [ ] `internal/ui/editor/host.go` package exists
- [ ] filebrowser plugin uses `editor.Host` for inline editing
- [ ] notes plugin uses `editor.Host` for inline editing
- [ ] Inline editing in filebrowser works identically (enter, type, save, exit)
- [ ] Inline editing in notes works identically (enter, type, save, exit)
- [ ] Exit confirmation dialog works in both plugins
- [ ] Click-away detection works in both plugins
- [ ] Mouse forwarding to tmux works
- [ ] `ConsumesTextInput()` returns true when editor is active
- [ ] Each plugin's inline editing setup code is < 30 lines

## Dependencies

- Uses `internal/tty` (existing)
- Uses `internal/modal` (existing)

## Risks

- **Medium**: filebrowser and notes may have subtle behavioral differences in editing
- **Mitigation**: Catalog all differences before creating the shared component, make them configurable via `Config`

## Scenarios

### Scenario: filebrowser enters edit mode
```
Given the user presses 'e' on a file in filebrowser
When the editor host is activated
Then the file content is loaded into tmux
And key presses are forwarded to the tmux pane
And the app footer shows "Save: Ctrl+S | Exit: Esc"
```

### Scenario: notes auto-save on exit
```
Given the user is editing a note
When they press Esc
Then the editor host triggers the OnSave callback
And transitions to inactive state
And the notes plugin shows the updated note preview
```

### Scenario: Unsaved changes confirmation
```
Given the user has modified content in the editor
When they press Esc
Then a "Save changes?" modal appears
When they choose "Discard"
Then the editor exits without saving
And OnExit is called with saved=false
```
