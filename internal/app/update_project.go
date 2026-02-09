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
	if m.projectAddMode {
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

	projects := m.projectSwitcherFiltered

	switch msg.Type {
	case tea.KeyEnter:
		// Select project and switch to it
		if m.projectSwitcherCursor >= 0 && m.projectSwitcherCursor < len(projects) {
			selectedProject := projects[m.projectSwitcherCursor]
			m.resetProjectSwitcher()
			m.updateContext()
			return m, m.switchProject(selectedProject.Path)
		}
		return m, nil

	case tea.KeyUp:
		m.projectSwitcherCursor--
		if m.projectSwitcherCursor < 0 {
			m.projectSwitcherCursor = 0
		}
		m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
		m.previewProjectTheme()
		return m, nil

	case tea.KeyDown:
		m.projectSwitcherCursor++
		if m.projectSwitcherCursor >= len(projects) {
			m.projectSwitcherCursor = len(projects) - 1
		}
		if m.projectSwitcherCursor < 0 {
			m.projectSwitcherCursor = 0
		}
		m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
		m.previewProjectTheme()
		return m, nil
	}

	// Handle non-text shortcuts
	switch msg.String() {
	case "ctrl+n":
		m.projectSwitcherCursor++
		if m.projectSwitcherCursor >= len(projects) {
			m.projectSwitcherCursor = len(projects) - 1
		}
		if m.projectSwitcherCursor < 0 {
			m.projectSwitcherCursor = 0
		}
		m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
		m.previewProjectTheme()
		return m, nil

	case "ctrl+p":
		m.projectSwitcherCursor--
		if m.projectSwitcherCursor < 0 {
			m.projectSwitcherCursor = 0
		}
		m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
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
	m.projectSwitcherInput, cmd = m.projectSwitcherInput.Update(msg)

	// Re-filter on input change
	m.projectSwitcherFiltered = filterProjects(allProjects, m.projectSwitcherInput.Value())
	m.clearProjectSwitcherModal() // Clear modal cache on filter change
	// Reset cursor if it's beyond filtered list
	if m.projectSwitcherCursor >= len(m.projectSwitcherFiltered) {
		m.projectSwitcherCursor = len(m.projectSwitcherFiltered) - 1
	}
	if m.projectSwitcherCursor < 0 {
		m.projectSwitcherCursor = 0
	}
	m.projectSwitcherScroll = 0
	m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
	m.previewProjectTheme()

	return m, cmd
}

// handleProjectSwitcherMouse handles mouse events for the project switcher modal.
func (m *Model) handleProjectSwitcherMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureProjectSwitcherModal()
	if m.projectSwitcherModal == nil {
		return m, nil
	}
	if m.projectSwitcherMouseHandler == nil {
		m.projectSwitcherMouseHandler = mouse.NewHandler()
	}

	// Handle scroll wheel for project list navigation
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.projectSwitcherCursor--
		if m.projectSwitcherCursor < 0 {
			m.projectSwitcherCursor = 0
		}
		m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(
			m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
		m.clearProjectSwitcherModal()
		m.previewProjectTheme()
		return m, nil
	case tea.MouseButtonWheelDown:
		projects := m.projectSwitcherFiltered
		m.projectSwitcherCursor++
		if m.projectSwitcherCursor >= len(projects) {
			m.projectSwitcherCursor = len(projects) - 1
		}
		if m.projectSwitcherCursor < 0 {
			m.projectSwitcherCursor = 0
		}
		m.projectSwitcherScroll = projectSwitcherEnsureCursorVisible(
			m.projectSwitcherCursor, m.projectSwitcherScroll, 8)
		m.clearProjectSwitcherModal()
		m.previewProjectTheme()
		return m, nil
	}

	action := m.projectSwitcherModal.HandleMouse(msg, m.projectSwitcherMouseHandler)

	// Check if action is a project item click
	if strings.HasPrefix(action, projectSwitcherItemPrefix) {
		// Extract index from item ID
		var idx int
		if _, err := fmt.Sscanf(action, projectSwitcherItemPrefix+"%d", &idx); err == nil {
			projects := m.projectSwitcherFiltered
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
		projects := m.projectSwitcherFiltered
		if m.projectSwitcherCursor >= 0 && m.projectSwitcherCursor < len(projects) {
			selectedProject := projects[m.projectSwitcherCursor]
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
	if m.projectAddCommunityMode {
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
		m.projectAddCommunityMode = true
		m.projectAddCommunityList = community.ListSchemes()
		m.projectAddCommunityCursor = 0
		m.projectAddCommunityScroll = 0
		return m, nil

	case "up", "k":
		if m.projectAddThemeCursor > 0 {
			m.projectAddThemeCursor--
			if m.projectAddThemeCursor < m.projectAddThemeScroll {
				m.projectAddThemeScroll = m.projectAddThemeCursor
			}
			m.previewProjectAddTheme()
		}
		return m, nil

	case "down", "j":
		if m.projectAddThemeCursor < len(m.projectAddThemeFiltered)-1 {
			m.projectAddThemeCursor++
			if m.projectAddThemeCursor >= m.projectAddThemeScroll+maxVisible {
				m.projectAddThemeScroll = m.projectAddThemeCursor - maxVisible + 1
			}
			m.previewProjectAddTheme()
		}
		return m, nil

	case "enter":
		if m.projectAddThemeCursor >= 0 && m.projectAddThemeCursor < len(m.projectAddThemeFiltered) {
			if m.projectAdd != nil {
				m.projectAdd.themeSelected = m.projectAddThemeFiltered[m.projectAddThemeCursor]
			}
		}
		m.projectAddModalWidth = 0 // Force modal rebuild to show new theme
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
	m.projectAddThemeInput, cmd = m.projectAddThemeInput.Update(msg)
	// Re-filter
	query := m.projectAddThemeInput.Value()
	all := append([]string{"(use global)"}, styles.ListThemes()...)
	if query == "" {
		m.projectAddThemeFiltered = all
	} else {
		var filtered []string
		q := strings.ToLower(query)
		for _, name := range all {
			if strings.Contains(strings.ToLower(name), q) {
				filtered = append(filtered, name)
			}
		}
		m.projectAddThemeFiltered = filtered
	}
	m.projectAddThemeCursor = 0
	m.projectAddThemeScroll = 0
	return m, cmd
}

// handleProjectAddCommunityKeys handles keys in the community sub-browser within add-project.
func (m *Model) handleProjectAddCommunityKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxVisible := 6
	switch msg.String() {
	case "esc", "tab":
		// Back to built-in themes
		m.projectAddCommunityMode = false
		// Restore theme
		resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
		return m, nil

	case "up", "k":
		if m.projectAddCommunityCursor > 0 {
			m.projectAddCommunityCursor--
			if m.projectAddCommunityCursor < m.projectAddCommunityScroll {
				m.projectAddCommunityScroll = m.projectAddCommunityCursor
			}
			m.previewProjectAddCommunity()
		}
		return m, nil

	case "down", "j":
		if m.projectAddCommunityCursor < len(m.projectAddCommunityList)-1 {
			m.projectAddCommunityCursor++
			if m.projectAddCommunityCursor >= m.projectAddCommunityScroll+maxVisible {
				m.projectAddCommunityScroll = m.projectAddCommunityCursor - maxVisible + 1
			}
			m.previewProjectAddCommunity()
		}
		return m, nil

	case "enter":
		if m.projectAddCommunityCursor >= 0 && m.projectAddCommunityCursor < len(m.projectAddCommunityList) {
			if m.projectAdd != nil {
				m.projectAdd.themeSelected = m.projectAddCommunityList[m.projectAddCommunityCursor]
			}
		}
		m.projectAddModalWidth = 0 // Force modal rebuild to show new theme
		m.resetProjectAddThemePicker()
		// Restore theme
		resolved := theme.ResolveTheme(m.cfg, m.ui.WorkDir)
		theme.ApplyResolved(resolved)
		return m, nil
	}

	return m, nil
}
