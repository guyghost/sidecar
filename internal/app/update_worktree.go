package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleWorktreeSwitcherKeys handles keyboard input for the worktree switcher modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleWorktreeSwitcherKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	worktrees := m.worktreeSwitcherFiltered

	switch msg.Type {
	case tea.KeyEnter:
		// Select worktree and switch to it
		if m.worktreeSwitcherCursor >= 0 && m.worktreeSwitcherCursor < len(worktrees) {
			selectedPath := worktrees[m.worktreeSwitcherCursor].Path
			m.resetWorktreeSwitcher()
			m.updateContext()
			return m, m.switchWorktree(selectedPath)
		}
		return m, nil

	case tea.KeyUp:
		m.worktreeSwitcherCursor--
		if m.worktreeSwitcherCursor < 0 {
			m.worktreeSwitcherCursor = 0
		}
		m.worktreeSwitcherScroll = worktreeSwitcherEnsureCursorVisible(m.worktreeSwitcherCursor, m.worktreeSwitcherScroll, 8)
		return m, nil

	case tea.KeyDown:
		m.worktreeSwitcherCursor++
		if m.worktreeSwitcherCursor >= len(worktrees) {
			m.worktreeSwitcherCursor = len(worktrees) - 1
		}
		if m.worktreeSwitcherCursor < 0 {
			m.worktreeSwitcherCursor = 0
		}
		m.worktreeSwitcherScroll = worktreeSwitcherEnsureCursorVisible(m.worktreeSwitcherCursor, m.worktreeSwitcherScroll, 8)
		return m, nil
	}

	// Handle non-text shortcuts
	switch msg.String() {
	case "ctrl+n":
		m.worktreeSwitcherCursor++
		if m.worktreeSwitcherCursor >= len(worktrees) {
			m.worktreeSwitcherCursor = len(worktrees) - 1
		}
		if m.worktreeSwitcherCursor < 0 {
			m.worktreeSwitcherCursor = 0
		}
		m.worktreeSwitcherScroll = worktreeSwitcherEnsureCursorVisible(m.worktreeSwitcherCursor, m.worktreeSwitcherScroll, 8)
		return m, nil

	case "ctrl+p":
		m.worktreeSwitcherCursor--
		if m.worktreeSwitcherCursor < 0 {
			m.worktreeSwitcherCursor = 0
		}
		m.worktreeSwitcherScroll = worktreeSwitcherEnsureCursorVisible(m.worktreeSwitcherCursor, m.worktreeSwitcherScroll, 8)
		return m, nil

	case "W":
		// Close modal with same key
		m.resetWorktreeSwitcher()
		m.updateContext()
		return m, nil
	}

	// Filter out unparsed mouse escape sequences
	if isMouseEscapeSequence(msg) {
		return m, nil
	}

	// Forward other keys to text input for filtering
	var cmd tea.Cmd
	m.worktreeSwitcherInput, cmd = m.worktreeSwitcherInput.Update(msg)

	// Re-filter on input change
	m.worktreeSwitcherFiltered = filterWorktrees(m.worktreeSwitcherAll, m.worktreeSwitcherInput.Value())
	m.clearWorktreeSwitcherModal() // Clear modal cache on filter change
	// Reset cursor if it's beyond filtered list
	if m.worktreeSwitcherCursor >= len(m.worktreeSwitcherFiltered) {
		m.worktreeSwitcherCursor = len(m.worktreeSwitcherFiltered) - 1
	}
	if m.worktreeSwitcherCursor < 0 {
		m.worktreeSwitcherCursor = 0
	}
	m.worktreeSwitcherScroll = 0
	m.worktreeSwitcherScroll = worktreeSwitcherEnsureCursorVisible(m.worktreeSwitcherCursor, m.worktreeSwitcherScroll, 8)

	return m, cmd
}
