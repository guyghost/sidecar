package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/config"
	"github.com/guyghost/sidecar/internal/mouse"
)

// handleThemeSwitcherKeys handles keyboard input for the theme switcher modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleThemeSwitcherKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ctrl+s or left/right toggles scope between global and project
	if m.currentProjectConfig() != nil {
		switch msg.String() {
		case "ctrl+s", "left", "right":
			if m.theme.Scope == "global" {
				m.theme.Scope = "project"
			} else {
				m.theme.Scope = "global"
			}
			return m, nil
		}
	}

	themes := m.theme.Filtered

	switch msg.Type {
	case tea.KeyEnter:
		// Confirm selection and close (ignore separators)
		if m.theme.SelectedIdx >= 0 && m.theme.SelectedIdx < len(themes) && !themes[m.theme.SelectedIdx].IsSeparator {
			entry := themes[m.theme.SelectedIdx]
			var tc config.ThemeConfig
			if entry.IsBuiltIn {
				tc = config.ThemeConfig{Name: entry.ThemeKey}
			} else {
				tc = config.ThemeConfig{Name: "default", Community: entry.ThemeKey}
			}
			m.previewThemeEntry(entry)
			return m, m.confirmThemeSelection(tc, entry.Name)
		}
		return m, nil

	case tea.KeyUp:
		m.theme.SelectedIdx--
		if m.theme.SelectedIdx < 0 {
			m.theme.SelectedIdx = 0
		}
		// Skip separators
		for m.theme.SelectedIdx > 0 && themes[m.theme.SelectedIdx].IsSeparator {
			m.theme.SelectedIdx--
		}
		if m.theme.SelectedIdx < len(themes) && !themes[m.theme.SelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.theme.SelectedIdx])
		}
		return m, nil

	case tea.KeyDown:
		m.theme.SelectedIdx++
		if m.theme.SelectedIdx >= len(themes) {
			m.theme.SelectedIdx = len(themes) - 1
		}
		if m.theme.SelectedIdx < 0 {
			m.theme.SelectedIdx = 0
		}
		// Skip separators
		for m.theme.SelectedIdx < len(themes)-1 && themes[m.theme.SelectedIdx].IsSeparator {
			m.theme.SelectedIdx++
		}
		if m.theme.SelectedIdx < len(themes) && !themes[m.theme.SelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.theme.SelectedIdx])
		}
		return m, nil
	}

	// Handle non-text shortcuts
	switch msg.String() {
	case "ctrl+n":
		m.theme.SelectedIdx++
		if m.theme.SelectedIdx >= len(themes) {
			m.theme.SelectedIdx = len(themes) - 1
		}
		if m.theme.SelectedIdx < 0 {
			m.theme.SelectedIdx = 0
		}
		for m.theme.SelectedIdx < len(themes)-1 && themes[m.theme.SelectedIdx].IsSeparator {
			m.theme.SelectedIdx++
		}
		if m.theme.SelectedIdx < len(themes) && !themes[m.theme.SelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.theme.SelectedIdx])
		}
		return m, nil

	case "ctrl+p":
		m.theme.SelectedIdx--
		if m.theme.SelectedIdx < 0 {
			m.theme.SelectedIdx = 0
		}
		for m.theme.SelectedIdx > 0 && themes[m.theme.SelectedIdx].IsSeparator {
			m.theme.SelectedIdx--
		}
		if m.theme.SelectedIdx < len(themes) && !themes[m.theme.SelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.theme.SelectedIdx])
		}
		return m, nil

	case "#":
		// Close modal and restore original
		m.previewThemeEntry(m.theme.Original)
		m.resetThemeSwitcher()
		m.updateContext()
		return m, nil
	}

	// Filter out unparsed mouse escape sequences
	if isMouseEscapeSequence(msg) {
		return m, nil
	}

	// Forward other keys to text input for filtering
	var cmd tea.Cmd
	m.theme.Input, cmd = m.theme.Input.Update(msg)

	// Re-filter on input change
	m.theme.Filtered = filterThemeEntries(buildUnifiedThemeList(), m.theme.Input.Value())
	m.clearThemeSwitcherModal() // Force modal rebuild
	if m.theme.SelectedIdx >= len(m.theme.Filtered) {
		m.theme.SelectedIdx = len(m.theme.Filtered) - 1
	}
	if m.theme.SelectedIdx < 0 {
		m.theme.SelectedIdx = 0
	}

	// Live preview current selection (skip separators)
	if m.theme.SelectedIdx >= 0 && m.theme.SelectedIdx < len(m.theme.Filtered) && !m.theme.Filtered[m.theme.SelectedIdx].IsSeparator {
		m.previewThemeEntry(m.theme.Filtered[m.theme.SelectedIdx])
	}

	return m, cmd
}

// handleThemeSwitcherMouse handles mouse events for the theme switcher modal.
func (m *Model) handleThemeSwitcherMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureThemeSwitcherModal()
	if m.theme.Modal == nil {
		return m, nil
	}
	if m.theme.MouseHandler == nil {
		m.theme.MouseHandler = mouse.NewHandler()
	}

	action := m.theme.Modal.HandleMouse(msg, m.theme.MouseHandler)
	switch action {
	case "select":
		themes := m.theme.Filtered
		if m.theme.SelectedIdx >= 0 && m.theme.SelectedIdx < len(themes) {
			entry := themes[m.theme.SelectedIdx]
			m.previewThemeEntry(entry)
			var tc config.ThemeConfig
			if entry.IsBuiltIn {
				tc = config.ThemeConfig{Name: entry.ThemeKey}
			} else {
				tc = config.ThemeConfig{Name: "default", Community: entry.ThemeKey}
			}
			return m, m.confirmThemeSelection(tc, entry.Name)
		}
	}
	return m, nil
}
