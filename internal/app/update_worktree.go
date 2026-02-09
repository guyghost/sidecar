package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleWorktreeSwitcherKeys handles keyboard input for the worktree switcher modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleWorktreeSwitcherKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	worktrees := m.worktree.Filtered

	switch msg.Type {
	case tea.KeyEnter:
		// Select worktree and switch to it
		if m.worktree.Cursor >= 0 && m.worktree.Cursor < len(worktrees) {
			selectedPath := worktrees[m.worktree.Cursor].Path
			m.resetWorktreeSwitcher()
			m.updateContext()
			return m, m.switchWorktree(selectedPath)
		}
		return m, nil

	case tea.KeyUp:
		m.worktree.Cursor--
		if m.worktree.Cursor < 0 {
			m.worktree.Cursor = 0
		}
		m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)
		return m, nil

	case tea.KeyDown:
		m.worktree.Cursor++
		if m.worktree.Cursor >= len(worktrees) {
			m.worktree.Cursor = len(worktrees) - 1
		}
		if m.worktree.Cursor < 0 {
			m.worktree.Cursor = 0
		}
		m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)
		return m, nil
	}

	// Handle non-text shortcuts
	switch msg.String() {
	case "ctrl+n":
		m.worktree.Cursor++
		if m.worktree.Cursor >= len(worktrees) {
			m.worktree.Cursor = len(worktrees) - 1
		}
		if m.worktree.Cursor < 0 {
			m.worktree.Cursor = 0
		}
		m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)
		return m, nil

	case "ctrl+p":
		m.worktree.Cursor--
		if m.worktree.Cursor < 0 {
			m.worktree.Cursor = 0
		}
		m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)
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
	m.worktree.Input, cmd = m.worktree.Input.Update(msg)

	// Re-filter on input change
	m.worktree.Filtered = filterWorktrees(m.worktree.All, m.worktree.Input.Value())
	m.clearWorktreeSwitcherModal() // Clear modal cache on filter change
	// Reset cursor if it's beyond filtered list
	if m.worktree.Cursor >= len(m.worktree.Filtered) {
		m.worktree.Cursor = len(m.worktree.Filtered) - 1
	}
	if m.worktree.Cursor < 0 {
		m.worktree.Cursor = 0
	}
	m.worktree.Scroll = 0
	m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)

	return m, cmd
}
