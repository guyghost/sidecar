package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/sidecar/internal/modal"
	"github.com/guyghost/sidecar/internal/mouse"
	"github.com/guyghost/sidecar/internal/styles"
	"github.com/guyghost/sidecar/internal/ui"
)

const (
	worktreeSwitcherFilterID   = "worktree-switcher-filter"
	worktreeSwitcherItemPrefix = "worktree-switcher-item-"
)

// worktreeSwitcherItemID returns the ID for a worktree item at the given index.
func worktreeSwitcherItemID(idx int) string {
	return fmt.Sprintf("%s%d", worktreeSwitcherItemPrefix, idx)
}

// initWorktreeSwitcher initializes the worktree switcher modal.
func (m *Model) initWorktreeSwitcher() {
	m.clearWorktreeSwitcherModal()

	ti := textinput.New()
	ti.Placeholder = "Filter worktrees..."
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40
	m.worktree.Input = ti

	// Load all worktrees
	m.worktree.All = GetWorktrees(m.ui.WorkDir)
	m.worktree.Filtered = m.worktree.All
	m.worktree.Cursor = 0
	m.worktree.Scroll = 0

	// Set cursor to current worktree if found
	for i, wt := range m.worktree.Filtered {
		normalizedPath, _ := normalizePath(wt.Path)
		normalizedWorkDir, _ := normalizePath(m.ui.WorkDir)
		if normalizedPath == normalizedWorkDir {
			m.worktree.Cursor = i
			break
		}
	}
}

// resetWorktreeSwitcher resets the worktree switcher modal state.
func (m *Model) resetWorktreeSwitcher() {
	m.worktree.Show = false
	m.worktree.Cursor = 0
	m.worktree.Scroll = 0
	m.worktree.Filtered = nil
	m.worktree.All = nil
	m.clearWorktreeSwitcherModal()
}

// clearWorktreeSwitcherModal clears the modal cache.
func (m *Model) clearWorktreeSwitcherModal() {
	m.worktree.Modal = nil
	m.worktree.ModalWidth = 0
	m.worktree.MouseHandler = nil
}

// filterWorktrees filters worktrees by branch name or path.
func filterWorktrees(all []WorktreeInfo, query string) []WorktreeInfo {
	if query == "" {
		return all
	}
	q := strings.ToLower(query)
	var matches []WorktreeInfo
	for _, wt := range all {
		if strings.Contains(strings.ToLower(wt.Branch), q) ||
			strings.Contains(strings.ToLower(filepath.Base(wt.Path)), q) {
			matches = append(matches, wt)
		}
	}
	return matches
}

// worktreeSwitcherEnsureCursorVisible adjusts scroll to keep cursor in view.
func worktreeSwitcherEnsureCursorVisible(cursor, scroll, maxVisible int) int {
	if cursor < scroll {
		return cursor
	}
	if cursor >= scroll+maxVisible {
		return cursor - maxVisible + 1
	}
	return scroll
}

// ensureWorktreeSwitcherModal builds/rebuilds the worktree switcher modal.
func (m *Model) ensureWorktreeSwitcherModal() {
	modalW := 60
	if modalW > m.width-4 {
		modalW = m.width - 4
	}
	if modalW < 30 {
		modalW = 30
	}

	// Only rebuild if modal doesn't exist or width changed
	if m.worktree.Modal != nil && m.worktree.ModalWidth == modalW {
		return
	}
	m.worktree.ModalWidth = modalW

	m.worktree.Modal = modal.New("Switch Worktree",
		modal.WithWidth(modalW),
		modal.WithHints(false),
	).
		AddSection(modal.Input(worktreeSwitcherFilterID, &m.worktree.Input, modal.WithSubmitOnEnter(false))).
		AddSection(m.worktreeSwitcherCountSection()).
		AddSection(modal.Spacer()).
		AddSection(m.worktreeSwitcherListSection()).
		AddSection(m.worktreeSwitcherHintsSection())
}

// worktreeSwitcherCountSection renders the worktree count.
func (m *Model) worktreeSwitcherCountSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		worktrees := m.worktree.Filtered
		allWorktrees := m.worktree.All

		var countText string
		if m.worktree.Input.Value() != "" {
			countText = fmt.Sprintf("%d of %d worktrees", len(worktrees), len(allWorktrees))
		} else if len(allWorktrees) > 0 {
			countText = fmt.Sprintf("%d worktrees", len(allWorktrees))
		}
		return modal.RenderedSection{Content: styles.Muted.Render(countText)}
	}, nil)
}

// worktreeSwitcherListSection renders the worktree list with selection.
func (m *Model) worktreeSwitcherListSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		worktrees := m.worktree.Filtered

		// No worktrees
		if len(worktrees) == 0 {
			return modal.RenderedSection{Content: styles.Muted.Render("No worktrees found")}
		}

		// Styles
		cursorStyle := lipgloss.NewStyle().Foreground(styles.Primary)
		nameNormalStyle := lipgloss.NewStyle().Foreground(styles.Secondary)
		nameSelectedStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
		nameCurrentStyle := lipgloss.NewStyle().Foreground(styles.Success).Bold(true)
		nameCurrentSelectedStyle := lipgloss.NewStyle().Foreground(styles.Success).Bold(true)
		mainBadgeStyle := lipgloss.NewStyle().Foreground(styles.Warning)

		// Determine current worktree
		normalizedWorkDir, _ := normalizePath(m.ui.WorkDir)

		maxVisible := 8
		visibleCount := len(worktrees)
		if visibleCount > maxVisible {
			visibleCount = maxVisible
		}
		scrollOffset := m.worktree.Scroll

		var sb strings.Builder
		focusables := make([]modal.FocusableInfo, 0, visibleCount)
		lineOffset := 0

		// Scroll indicator (top)
		if scrollOffset > 0 {
			sb.WriteString(styles.Muted.Render(fmt.Sprintf("  ↑ %d more above", scrollOffset)))
			sb.WriteString("\n")
			lineOffset++
		}

		for i := scrollOffset; i < scrollOffset+visibleCount && i < len(worktrees); i++ {
			wt := worktrees[i]
			isCursor := i == m.worktree.Cursor
			itemID := worktreeSwitcherItemID(i)
			isHovered := itemID == hoverID

			normalizedPath, _ := normalizePath(wt.Path)
			isCurrent := normalizedPath == normalizedWorkDir

			// Cursor indicator
			if isCursor {
				sb.WriteString(cursorStyle.Render("> "))
			} else {
				sb.WriteString("  ")
			}

			// Determine display name (branch name for worktrees, "main" badge for main repo)
			displayName := wt.Branch
			if displayName == "" {
				displayName = filepath.Base(wt.Path)
			}

			// Name styling
			var nameStyle lipgloss.Style
			if isCurrent {
				if isCursor || isHovered {
					nameStyle = nameCurrentSelectedStyle
				} else {
					nameStyle = nameCurrentStyle
				}
			} else if isCursor || isHovered {
				nameStyle = nameSelectedStyle
			} else {
				nameStyle = nameNormalStyle
			}

			sb.WriteString(nameStyle.Render(displayName))

			// Main badge
			if wt.IsMain {
				sb.WriteString(" ")
				sb.WriteString(mainBadgeStyle.Render("[main]"))
			}

			// Current indicator
			if isCurrent {
				sb.WriteString(styles.Muted.Render(" (current)"))
			}

			sb.WriteString("\n")

			// Show path (truncated if needed)
			pathDisplay := wt.Path
			maxPathLen := contentWidth - 4
			if len(pathDisplay) > maxPathLen {
				pathDisplay = "..." + pathDisplay[len(pathDisplay)-maxPathLen+3:]
			}
			sb.WriteString(styles.Muted.Render("  " + pathDisplay))

			if i < scrollOffset+visibleCount-1 && i < len(worktrees)-1 {
				sb.WriteString("\n")
			}

			// Each worktree takes 2 lines (name + path)
			focusables = append(focusables, modal.FocusableInfo{
				ID:      itemID,
				OffsetX: 0,
				OffsetY: lineOffset + (i-scrollOffset)*2,
				Width:   contentWidth,
				Height:  2,
			})
		}

		// Scroll indicator (bottom)
		remaining := len(worktrees) - (scrollOffset + visibleCount)
		if remaining > 0 {
			sb.WriteString("\n")
			sb.WriteString(styles.Muted.Render(fmt.Sprintf("  ↓ %d more below", remaining)))
		}

		return modal.RenderedSection{Content: sb.String(), Focusables: focusables}
	}, m.worktreeSwitcherListUpdate)
}

// worktreeSwitcherListUpdate handles key events for the worktree list.
func (m *Model) worktreeSwitcherListUpdate(msg tea.Msg, focusID string) (string, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return "", nil
	}

	worktrees := m.worktree.Filtered
	if len(worktrees) == 0 {
		return "", nil
	}

	switch keyMsg.String() {
	case "up", "k", "ctrl+p":
		if m.worktree.Cursor > 0 {
			m.worktree.Cursor--
			m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)
			m.worktree.ModalWidth = 0 // Force modal rebuild for scroll
		}
		return "", nil

	case "down", "j", "ctrl+n":
		if m.worktree.Cursor < len(worktrees)-1 {
			m.worktree.Cursor++
			m.worktree.Scroll = worktreeSwitcherEnsureCursorVisible(m.worktree.Cursor, m.worktree.Scroll, 8)
			m.worktree.ModalWidth = 0 // Force modal rebuild for scroll
		}
		return "", nil

	case "enter":
		if m.worktree.Cursor >= 0 && m.worktree.Cursor < len(worktrees) {
			return "select", nil
		}
		return "", nil
	}

	return "", nil
}

// worktreeSwitcherHintsSection renders the help text.
func (m *Model) worktreeSwitcherHintsSection() modal.Section {
	return modal.Custom(func(contentWidth int, focusID, hoverID string) modal.RenderedSection {
		worktrees := m.worktree.Filtered

		var sb strings.Builder
		sb.WriteString("\n")

		if len(worktrees) == 0 {
			sb.WriteString(styles.KeyHint.Render("esc"))
			sb.WriteString(styles.Muted.Render(" clear filter  "))
			sb.WriteString(styles.KeyHint.Render("W"))
			sb.WriteString(styles.Muted.Render(" close"))
		} else {
			sb.WriteString(styles.KeyHint.Render("enter"))
			sb.WriteString(styles.Muted.Render(" switch  "))
			sb.WriteString(styles.KeyHint.Render("↑/↓"))
			sb.WriteString(styles.Muted.Render(" navigate  "))
			sb.WriteString(styles.KeyHint.Render("esc"))
			sb.WriteString(styles.Muted.Render(" cancel"))
		}

		return modal.RenderedSection{Content: sb.String()}
	}, nil)
}

// renderWorktreeSwitcherModal renders the worktree switcher modal.
func (m *Model) renderWorktreeSwitcherModal(content string) string {
	m.ensureWorktreeSwitcherModal()
	if m.worktree.Modal == nil {
		return content
	}

	if m.worktree.MouseHandler == nil {
		m.worktree.MouseHandler = mouse.NewHandler()
	}
	modalContent := m.worktree.Modal.Render(m.width, m.height, m.worktree.MouseHandler)
	return ui.OverlayModal(content, modalContent, m.width, m.height)
}

// handleWorktreeSwitcherMouse handles mouse events for the worktree switcher modal.
func (m *Model) handleWorktreeSwitcherMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureWorktreeSwitcherModal()
	if m.worktree.Modal == nil {
		return m, nil
	}
	if m.worktree.MouseHandler == nil {
		m.worktree.MouseHandler = mouse.NewHandler()
	}

	action := m.worktree.Modal.HandleMouse(msg, m.worktree.MouseHandler)

	// Check if action is a worktree item click
	if strings.HasPrefix(action, worktreeSwitcherItemPrefix) {
		var idx int
		if _, err := fmt.Sscanf(action, worktreeSwitcherItemPrefix+"%d", &idx); err == nil {
			worktrees := m.worktree.Filtered
			if idx >= 0 && idx < len(worktrees) {
				selectedPath := worktrees[idx].Path
				m.resetWorktreeSwitcher()
				m.updateContext()
				return m, m.switchWorktree(selectedPath)
			}
		}
		return m, nil
	}

	switch action {
	case "cancel":
		m.resetWorktreeSwitcher()
		m.updateContext()
		return m, nil
	case "select":
		worktrees := m.worktree.Filtered
		if m.worktree.Cursor >= 0 && m.worktree.Cursor < len(worktrees) {
			selectedPath := worktrees[m.worktree.Cursor].Path
			m.resetWorktreeSwitcher()
			m.updateContext()
			return m, m.switchWorktree(selectedPath)
		}
		return m, nil
	}

	return m, nil
}

// switchWorktree switches all plugins to a new worktree directory.
func (m *Model) switchWorktree(worktreePath string) tea.Cmd {
	// Skip if already on this worktree
	normalizedPath, _ := normalizePath(worktreePath)
	normalizedWorkDir, _ := normalizePath(m.ui.WorkDir)
	if normalizedPath == normalizedWorkDir {
		return func() tea.Msg {
			return ToastMsg{Message: "Already on this worktree", Duration: 2 * time.Second}
		}
	}

	// Validate that the worktree still exists before switching
	if !WorktreeExists(worktreePath) {
		return func() tea.Msg {
			return ToastMsg{Message: "Worktree no longer exists", Duration: 3 * time.Second, IsError: true}
		}
	}

	// Use the same switchProject mechanism - it handles reinit, state save/restore
	return m.switchProject(worktreePath)
}


// refreshWorktreeCache calls GetWorktrees and caches the result for the current WorkDir.
func (m *Model) refreshWorktreeCache() {
	worktrees := GetWorktrees(m.ui.WorkDir)
	normalizedWorkDir, _ := normalizePath(m.ui.WorkDir)
	m.worktree.CachedInfo = nil
	for i, wt := range worktrees {
		normalizedPath, _ := normalizePath(wt.Path)
		if normalizedPath == normalizedWorkDir {
			m.worktree.CachedInfo = &worktrees[i]
			return
		}
	}
}

// currentWorktreeInfo returns the cached WorktreeInfo for the current WorkDir, or nil.
// Cache is populated eagerly in Update() (TickMsg, switchProject) — never in View().
func (m *Model) currentWorktreeInfo() *WorktreeInfo {
	return m.worktree.CachedInfo
}
