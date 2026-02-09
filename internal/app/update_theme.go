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
			if m.themeSwitcherScope == "global" {
				m.themeSwitcherScope = "project"
			} else {
				m.themeSwitcherScope = "global"
			}
			return m, nil
		}
	}

	themes := m.themeSwitcherFiltered

	switch msg.Type {
	case tea.KeyEnter:
		// Confirm selection and close (ignore separators)
		if m.themeSwitcherSelectedIdx >= 0 && m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
			entry := themes[m.themeSwitcherSelectedIdx]
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
		m.themeSwitcherSelectedIdx--
		if m.themeSwitcherSelectedIdx < 0 {
			m.themeSwitcherSelectedIdx = 0
		}
		// Skip separators
		for m.themeSwitcherSelectedIdx > 0 && themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.themeSwitcherSelectedIdx--
		}
		if m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.themeSwitcherSelectedIdx])
		}
		return m, nil

	case tea.KeyDown:
		m.themeSwitcherSelectedIdx++
		if m.themeSwitcherSelectedIdx >= len(themes) {
			m.themeSwitcherSelectedIdx = len(themes) - 1
		}
		if m.themeSwitcherSelectedIdx < 0 {
			m.themeSwitcherSelectedIdx = 0
		}
		// Skip separators
		for m.themeSwitcherSelectedIdx < len(themes)-1 && themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.themeSwitcherSelectedIdx++
		}
		if m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.themeSwitcherSelectedIdx])
		}
		return m, nil
	}

	// Handle non-text shortcuts
	switch msg.String() {
	case "ctrl+n":
		m.themeSwitcherSelectedIdx++
		if m.themeSwitcherSelectedIdx >= len(themes) {
			m.themeSwitcherSelectedIdx = len(themes) - 1
		}
		if m.themeSwitcherSelectedIdx < 0 {
			m.themeSwitcherSelectedIdx = 0
		}
		for m.themeSwitcherSelectedIdx < len(themes)-1 && themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.themeSwitcherSelectedIdx++
		}
		if m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.themeSwitcherSelectedIdx])
		}
		return m, nil

	case "ctrl+p":
		m.themeSwitcherSelectedIdx--
		if m.themeSwitcherSelectedIdx < 0 {
			m.themeSwitcherSelectedIdx = 0
		}
		for m.themeSwitcherSelectedIdx > 0 && themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.themeSwitcherSelectedIdx--
		}
		if m.themeSwitcherSelectedIdx < len(themes) && !themes[m.themeSwitcherSelectedIdx].IsSeparator {
			m.previewThemeEntry(themes[m.themeSwitcherSelectedIdx])
		}
		return m, nil

	case "#":
		// Close modal and restore original
		m.previewThemeEntry(m.themeSwitcherOriginal)
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
	m.themeSwitcherInput, cmd = m.themeSwitcherInput.Update(msg)

	// Re-filter on input change
	m.themeSwitcherFiltered = filterThemeEntries(buildUnifiedThemeList(), m.themeSwitcherInput.Value())
	m.clearThemeSwitcherModal() // Force modal rebuild
	if m.themeSwitcherSelectedIdx >= len(m.themeSwitcherFiltered) {
		m.themeSwitcherSelectedIdx = len(m.themeSwitcherFiltered) - 1
	}
	if m.themeSwitcherSelectedIdx < 0 {
		m.themeSwitcherSelectedIdx = 0
	}

	// Live preview current selection (skip separators)
	if m.themeSwitcherSelectedIdx >= 0 && m.themeSwitcherSelectedIdx < len(m.themeSwitcherFiltered) && !m.themeSwitcherFiltered[m.themeSwitcherSelectedIdx].IsSeparator {
		m.previewThemeEntry(m.themeSwitcherFiltered[m.themeSwitcherSelectedIdx])
	}

	return m, cmd
}

// handleThemeSwitcherMouse handles mouse events for the theme switcher modal.
func (m *Model) handleThemeSwitcherMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureThemeSwitcherModal()
	if m.themeSwitcherModal == nil {
		return m, nil
	}
	if m.themeSwitcherMouseHandler == nil {
		m.themeSwitcherMouseHandler = mouse.NewHandler()
	}

	action := m.themeSwitcherModal.HandleMouse(msg, m.themeSwitcherMouseHandler)
	switch action {
	case "select":
		themes := m.themeSwitcherFiltered
		if m.themeSwitcherSelectedIdx >= 0 && m.themeSwitcherSelectedIdx < len(themes) {
			entry := themes[m.themeSwitcherSelectedIdx]
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
