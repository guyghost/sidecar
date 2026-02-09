package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/community"
	"github.com/guyghost/sidecar/internal/mouse"
	"github.com/guyghost/sidecar/internal/styles"
	"github.com/guyghost/sidecar/internal/theme"
)

// handleProjectSwitcherKeys handles keyboard input for the project switcher modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleProjectSwitcherKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle project add sub-mode keys
	if m.project.AddMode {
		return m.handleProjectAddModalKeys(msg)
	}

	allProjects := m.cfg.Projects.List
	if len(allProjects) == 0 {
		// No projects configured - handle y for LLM prompt, ctrl+a for add, close on q/@
		switch msg.String() {
		case "y":
			return m, m.copyProjectSetupPrompt()
		case "ctrl+a":
			m.initProjectAdd()
			return m, nil
		case "q", "@":
			m.resetProjectSwitcher()
			m.updateContext()
		}
		return m, nil
	}

	projects := m.project.Filtered

	switch msg.Type {
	case tea.KeyEnter:
		// Select project and switch to it
		if m.project.Cursor >= 0 && m.project.Cursor < len(projects) {
			selectedProject := projects[m.project.Cursor]
			m.resetProjectSwitcher()
			m.updateContext()
			return m, m.switchProject(selectedProject.Path)
		}
		return m, nil

	case tea.KeyUp:
		m.project.Cursor--
		if m.project.Cursor < 0 {
			m.project.Cursor = 0
		}
		m.project.Scroll = projectSwitcherEnsureCursorVisible(m.project.Cursor, m.project.Scroll, 8)
		m.previewProjectTheme()
		return m, nil

	case tea.KeyDown:
		m.project.Cursor++
		if m.project.Cursor >= len(projects) {
			m.project.Cursor = len(projects) - 1
		}
		if m.project.Cursor < 0 {
			m.project.Cursor = 0
		}
		m.project.Scroll = projectSwitcherEnsureCursorVisible(m.project.Cursor, m.project.Scroll, 8)
		m.previewProjectTheme()
		return m, nil
	}

	// Handle non-text shortcuts
	switch msg.String() {
	case "ctrl+n":
		m.project.Cursor++
		if m.project.Cursor >= len(projects) {
			m.project.Cursor = len(projects) - 1
		}
		if m.project.Cursor < 0 {
			m.project.Cursor = 0
		}
		m.project.Scroll = projectSwitcherEnsureCursorVisible(m.project.Cursor, m.project.Scroll, 8)
		m.previewProjectTheme()
		return m, nil

	case "ctrl+p":
		m.project.Cursor--
		if m.project.Cursor < 0 {
			m.project.Cursor = 0
		}
		m.project.Scroll = projectSwitcherEnsureCursorVisible(m.project.Cursor, m.project.Scroll, 8)
		m.previewProjectTheme()
		return m, nil

	case "ctrl+a":
		m.initProjectAdd()
		return m, nil

	case "@":
		// Close modal
		m.resetProjectSwitcher()
		m.updateContext()
		return m, nil
	}

	// Filter out unparsed mouse escape sequences
	if isMouseEscapeSequence(msg) {
		return m, nil
	}

	// Forward other keys to text input for filtering
	var cmd tea.Cmd
	m.project.Input, cmd = m.project.Input.Update(msg)

	// Re-filter on input change
	m.project.Filtered = filterProjects(allProjects, m.project.Input.Value())
	m.clearProjectSwitcherModal() // Clear modal cache on filter change
	// Reset cursor if it's beyond filtered list
	if m.project.Cursor >= len(m.project.Filtered) {
		m.project.Cursor = len(m.project.Filtered) - 1
	}
	if m.project.Cursor < 0 {
		m.project.Cursor = 0
	}
	m.project.Scroll = 0
	m.project.Scroll = projectSwitcherEnsureCursorVisible(m.project.Cursor, m.project.Scroll, 8)
	m.previewProjectTheme()

	return m, cmd
}

// handleProjectSwitcherMouse handles mouse events for the project switcher modal.
func (m *Model) handleProjectSwitcherMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureProjectSwitcherModal()
	if m.project.Modal == nil {
		return m, nil
	}
	if m.project.MouseHandler == nil {
		m.project.MouseHandler = mouse.NewHandler()
	}

	// Handle scroll wheel for project list navigation
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.project.Cursor--
		if m.project.Cursor < 0 {
			m.project.Cursor = 0
		}
		m.project.Scroll = projectSwitcherEnsureCursorVisible(
			m.project.Cursor, m.project.Scroll, 8)
		m.clearProjectSwitcherModal()
		m.previewProjectTheme()
		return m, nil
	case tea.MouseButtonWheelDown:
		projects := m.project.Filtered
		m.project.Cursor++
		if m.project.Cursor >= len(projects) {
			m.project.Cursor = len(projects) - 1
		}
		if m.project.Cursor < 0 {
			m.project.Cursor = 0
		}
		m.project.Scroll = projectSwitcherEnsureCursorVisible(
			m.project.Cursor, m.project.Scroll, 8)
		m.clearProjectSwitcherModal()
		m.previewProjectTheme()
		return m, nil
	}

	action := m.project.Modal.HandleMouse(msg, m.project.MouseHandler)

	// Check if action is a project item click
	if strings.HasPrefix(action, projectSwitcherItemPrefix) {
		// Extract index from item ID
		var idx int
		if _, err := fmt.Sscanf(action, projectSwitcherItemPrefix+"%d", &idx); err == nil {
			projects := m.project.Filtered
			if idx >= 0 && idx < len(projects) {
				selectedProject := projects[idx]
				m.resetProjectSwitcher()
				m.updateContext()
				return m, m.switchProject(selectedProject.Path)
			}
		}
		return m, nil
	}

	switch action {
	case "cancel":
		m.resetProjectSwitcher()
		m.updateContext()
		return m, nil
	case "select":
		projects := m.project.Filtered
		if m.project.Cursor >= 0 && m.project.Cursor < len(projects) {
			selectedProject := projects[m.project.Cursor]
			m.resetProjectSwitcher()
			m.updateContext()
			return m, m.switchProject(selectedProject.Path)
		}
		return m, nil
	}

	return m, nil
}

// handleProjectAddThemePickerKeys handles keys within the theme picker sub-modal.
func (m *Model) handleProjectAddThemePickerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.project.AddCommunityMode {
		return m.handleProjectAddCommunityKeys(msg)
	}

	maxVisible := 6
	switch msg.String() {
	case "esc":
		m.resetProjectAddThemePicker()
		// Restore theme
		resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
		return m, nil

	case "tab":
		// Switch to community themes
		m.project.AddCommunityMode = true
		m.project.AddCommunityList = community.ListSchemes()
		m.project.AddCommunityCursor = 0
		m.project.AddCommunityScroll = 0
		return m, nil

	case "up", "k":
		if m.project.AddThemeCursor > 0 {
			m.project.AddThemeCursor--
			if m.project.AddThemeCursor < m.project.AddThemeScroll {
				m.project.AddThemeScroll = m.project.AddThemeCursor
			}
			m.previewProjectAddTheme()
		}
		return m, nil

	case "down", "j":
		if m.project.AddThemeCursor < len(m.project.AddThemeFiltered)-1 {
			m.project.AddThemeCursor++
			if m.project.AddThemeCursor >= m.project.AddThemeScroll+maxVisible {
				m.project.AddThemeScroll = m.project.AddThemeCursor - maxVisible + 1
			}
			m.previewProjectAddTheme()
		}
		return m, nil

	case "enter":
		if m.project.AddThemeCursor >= 0 && m.project.AddThemeCursor < len(m.project.AddThemeFiltered) {
			if m.project.Add != nil {
				m.project.Add.themeSelected = m.project.AddThemeFiltered[m.project.AddThemeCursor]
			}
		}
		m.project.AddModalWidth = 0 // Force modal rebuild to show new theme
		m.resetProjectAddThemePicker()
		// Restore theme
		resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
		return m, nil
	}

	// Filter out unparsed mouse escape sequences
	if isMouseEscapeSequence(msg) {
		return m, nil
	}

	// Forward to filter input
	var cmd tea.Cmd
	m.project.AddThemeInput, cmd = m.project.AddThemeInput.Update(msg)
	// Re-filter
	query := m.project.AddThemeInput.Value()
	all := append([]string{"(use global)"}, styles.ListThemes()...)
	if query == "" {
		m.project.AddThemeFiltered = all
	} else {
		var filtered []string
		q := strings.ToLower(query)
		for _, name := range all {
			if strings.Contains(strings.ToLower(name), q) {
				filtered = append(filtered, name)
			}
		}
		m.project.AddThemeFiltered = filtered
	}
	m.project.AddThemeCursor = 0
	m.project.AddThemeScroll = 0
	return m, cmd
}

// handleProjectAddCommunityKeys handles keys in the community sub-browser within add-project.
func (m *Model) handleProjectAddCommunityKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxVisible := 6
	switch msg.String() {
	case "esc", "tab":
		// Back to built-in themes
		m.project.AddCommunityMode = false
		// Restore theme
		resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
		return m, nil

	case "up", "k":
		if m.project.AddCommunityCursor > 0 {
			m.project.AddCommunityCursor--
			if m.project.AddCommunityCursor < m.project.AddCommunityScroll {
				m.project.AddCommunityScroll = m.project.AddCommunityCursor
			}
			m.previewProjectAddCommunity()
		}
		return m, nil

	case "down", "j":
		if m.project.AddCommunityCursor < len(m.project.AddCommunityList)-1 {
			m.project.AddCommunityCursor++
			if m.project.AddCommunityCursor >= m.project.AddCommunityScroll+maxVisible {
				m.project.AddCommunityScroll = m.project.AddCommunityCursor - maxVisible + 1
			}
			m.previewProjectAddCommunity()
		}
		return m, nil

	case "enter":
		if m.project.AddCommunityCursor >= 0 && m.project.AddCommunityCursor < len(m.project.AddCommunityList) {
			if m.project.Add != nil {
				m.project.Add.themeSelected = m.project.AddCommunityList[m.project.AddCommunityCursor]
			}
		}
		m.project.AddModalWidth = 0 // Force modal rebuild to show new theme
		m.resetProjectAddThemePicker()
		// Restore theme
		resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
		return m, nil
	}

	return m, nil
}
