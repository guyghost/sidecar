package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/version"
)

// handlePluginSwitchKeys handles backtick, tilde, and number keys for plugin switching.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handlePluginSwitchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "`":
		// Backtick cycles to next plugin (except in text input contexts)
		if m.consumesTextInput() {
			return m, nil, false
		}
		return m, m.NextPlugin(), true
	case "~":
		// Tilde cycles to previous plugin (except in text input contexts)
		if m.consumesTextInput() {
			return m, nil, false
		}
		return m, m.PrevPlugin(), true
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		// Number keys for direct plugin switching
		// Block in text input contexts (user is typing numbers)
		if m.consumesTextInput() {
			return m, nil, false
		}
		idx := int(msg.Runes[0] - '1')
		return m, m.SetActivePlugin(idx), true
	}
	return m, nil, false
}

// handleToggleKeys handles toggle shortcuts (?, !, @, W, #, i, r).
// Returns (model, cmd, handled). If handled is false, the key should fall through.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleToggleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch msg.String() {
	case "?":
		m.showPalette = !m.showPalette
		if m.showPalette {
			// Open palette with current context
			pluginCtx := "global"
			if p := m.ActivePlugin(); p != nil {
				pluginCtx = p.ID()
			}
			m.palette.SetSize(m.width, m.height)
			m.palette.Open(m.keymap, m.registry.Plugins(), m.activeContext, pluginCtx)
			m.activeContext = "palette"
		} else {
			m.updateContext()
		}
		return m, nil, true
	case "!":
		m.showDiagnostics = !m.showDiagnostics
		if m.showDiagnostics {
			m.activeContext = "diagnostics"
			// Force version check in background (bypasses cache)
			return m, tea.Batch(
				version.ForceCheckAsync(m.currentVersion),
				version.ForceCheckTdAsync(),
			), true
		}
		m.clearDiagnosticsModal()
		m.updateContext()
		return m, nil, true
	case "@":
		// Toggle project switcher modal
		m.showProjectSwitcher = !m.showProjectSwitcher
		if m.showProjectSwitcher {
			m.activeContext = "project-switcher"
			m.initProjectSwitcher()
		} else {
			m.resetProjectSwitcher()
			m.updateContext()
		}
		return m, nil, true
	case "W":
		// Toggle worktree switcher modal (capital W)
		// Only enable if we're in a git repo with worktrees
		worktrees := GetWorktrees(m.ui.WorkDir)
		if len(worktrees) <= 1 {
			// No worktrees or only main repo - show toast
			return m, func() tea.Msg {
				return ToastMsg{Message: "No worktrees found", Duration: 2 * time.Second}
			}, true
		}
		m.showWorktreeSwitcher = !m.showWorktreeSwitcher
		if m.showWorktreeSwitcher {
			m.activeContext = "worktree-switcher"
			m.initWorktreeSwitcher()
		} else {
			m.resetWorktreeSwitcher()
			m.updateContext()
		}
		return m, nil, true
	case "#":
		// Toggle theme switcher modal
		m.showThemeSwitcher = !m.showThemeSwitcher
		if m.showThemeSwitcher {
			m.activeContext = "theme-switcher"
			m.initThemeSwitcher()
		} else {
			m.previewThemeEntry(m.themeSwitcherOriginal)
			m.resetThemeSwitcher()
			m.updateContext()
		}
		return m, nil, true
	case "i":
		if !m.hasModal() {
			m.showIssueInput = true
			m.activeContext = "issue-input"
			m.initIssueInput()
			return m, nil, true
		}
	case "r":
		// Forward 'r' to plugin in contexts where it's used for specific actions
		// or where the user is typing text input
		if !isGlobalRefreshContext(m.activeContext) {
			// Fall through to forward to plugin
			return m, nil, false
		}
		return m, Refresh(), true
	}
	return m, nil, false
}

// handleKeymapAndPluginForward tries keymap bindings, then forwards to active plugin.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleKeymapAndPluginForward(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Try keymap for context-specific bindings
	if cmd := m.keymap.Handle(msg, m.activeContext); cmd != nil {
		return m, cmd
	}

	// Forward to active plugin
	if p := m.ActivePlugin(); p != nil {
		newPlugin, cmd := p.Update(msg)
		plugins := m.registry.Plugins()
		if m.activePlugin < len(plugins) {
			plugins[m.activePlugin] = newPlugin
		}
		m.updateContext()
		return m, cmd
	}

	return m, nil
}
