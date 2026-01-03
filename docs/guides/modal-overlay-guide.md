# Modal Overlay Implementation Guide

This guide covers how to implement modals with dimmed backgrounds in Sidecar.

## Overview

Modals should dim the background to:
- Draw user focus to the modal content
- Provide visual separation between modal and underlying content
- Create a consistent, polished UX across the application

## Two Approaches

### 1. App-Level Modals (Full-Screen Control)

For modals rendered at the app level (`internal/app/view.go`), use `lipgloss.Place()` with whitespace options:

```go
func (m Model) renderMyModal(content string) string {
    modal := styles.ModalBox.Render(content)

    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        modal,
        lipgloss.WithWhitespaceChars(" "),
        lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
    )
}
```

**How it works:**
- `lipgloss.Place()` centers the modal and fills surrounding space with whitespace
- `WithWhitespaceChars(" ")` uses spaces as the fill character
- `WithWhitespaceForeground("#000000")` colors those spaces black, creating a dim effect

**Examples:** Help modal, Command palette

### 2. Plugin-Level Modals (Within Plugin Bounds)

For modals rendered within plugins, use the `overlayModal()` function from the gitstatus plugin:

```go
// In your plugin's render function:
func (p *Plugin) renderMyModal() string {
    // Render what should appear behind the modal
    background := p.renderNormalView()

    // Render your modal content with border
    modalContent := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(styles.Primary).
        Padding(1, 2).
        Width(modalWidth).
        Render(content)

    // Overlay modal on dimmed background
    return overlayModal(background, modalContent, p.width, p.height)
}
```

**How `overlayModal()` works** (from `internal/plugins/gitstatus/push_menu.go`):
1. Splits background into lines
2. Strips ANSI codes from each line and applies dim style (color 240)
3. Centers modal vertically and horizontally
4. Composites: dimmed background above + modal lines + dimmed background below

**Examples:** Git commit modal, Push menu

## Implementation Checklist

When adding dimmed background to a modal:

1. **Identify the modal type:**
   - App-level (full screen) → Use `lipgloss.Place()` with whitespace options
   - Plugin-level (bounded) → Use `overlayModal()` helper

2. **For app-level modals**, add these options to `lipgloss.Place()`:
   ```go
   lipgloss.WithWhitespaceChars(" "),
   lipgloss.WithWhitespaceForeground(lipgloss.Color("#000000")),
   ```

3. **For plugin-level modals:**
   - Import or copy the `overlayModal()` function
   - Ensure you have access to `ansi.Strip()` from `github.com/charmbracelet/x/ansi`
   - Pass raw modal content (don't pre-center with `lipgloss.Place()`)

## Style Constants

```go
// Dim style for background (plugin-level)
var dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// App-level dim color
const dimColor = "#000000"
```

## Common Pitfalls

1. **Don't use `lipgloss.Place()` with `overlayModal()`** - they both handle centering, which causes layout issues

2. **ANSI code handling** - When dimming, strip ANSI codes first to avoid corrupted output:
   ```go
   stripped := ansi.Strip(line)
   dimmed := dimStyle.Render(stripped)
   ```

3. **Height constraints** - Ensure modal content respects available height to prevent overflow

## File Locations

- App-level modals: `internal/app/view.go`
- Plugin modal helper: `internal/plugins/gitstatus/push_menu.go` (`overlayModal()`)
- Modal styles: `internal/styles/styles.go` (`ModalBox`, `ModalTitle`, etc.)
