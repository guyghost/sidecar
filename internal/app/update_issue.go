package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/mouse"
)

// handleIssueInputKeys handles keyboard input for the issue input modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleIssueInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ctrl+x toggles closed issue visibility (before type switch)
	if msg.String() == "ctrl+x" {
		m.issueSearchIncludeClosed = !m.issueSearchIncludeClosed
		m.issueSearchScrollOffset = 0
		m.issueSearchCursor = -1
		m.issueInputModal = nil
		m.issueInputModalWidth = 0
		if len(strings.TrimSpace(m.issueInputInput.Value())) >= 2 {
			m.issueSearchLoading = true
			return m, issueSearchCmd(m.ui.WorkDir, strings.TrimSpace(m.issueInputInput.Value()), m.issueSearchIncludeClosed)
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		return m.issueInputSubmit()
	case tea.KeyUp:
		if len(m.issueSearchResults) > 0 {
			m.issueSearchCursor--
			if m.issueSearchCursor < -1 {
				m.issueSearchCursor = -1
			}
			// Keep cursor visible in viewport
			if m.issueSearchCursor >= 0 && m.issueSearchCursor < m.issueSearchScrollOffset {
				m.issueSearchScrollOffset = m.issueSearchCursor
			}
			m.issueInputModal = nil
			m.issueInputModalWidth = 0
			return m, nil
		}
	case tea.KeyDown:
		if len(m.issueSearchResults) > 0 {
			m.issueSearchCursor++
			if m.issueSearchCursor >= len(m.issueSearchResults) {
				m.issueSearchCursor = len(m.issueSearchResults) - 1
			}
			// Keep cursor visible in viewport
			const maxVisible = 10
			if m.issueSearchCursor >= m.issueSearchScrollOffset+maxVisible {
				m.issueSearchScrollOffset = m.issueSearchCursor - maxVisible + 1
			}
			m.issueInputModal = nil
			m.issueInputModalWidth = 0
			return m, nil
		}
	case tea.KeyTab:
		if m.issueSearchCursor >= 0 && m.issueSearchCursor < len(m.issueSearchResults) {
			m.issueInputInput.SetValue(m.issueSearchResults[m.issueSearchCursor].ID)
			m.issueInputInput.CursorEnd()
			m.issueInputModal = nil
			m.issueInputModalWidth = 0
		}
		// Tab is consumed (fill-in or no-op) — don't forward to textinput
		return m, nil
	}

	if isMouseEscapeSequence(msg) {
		return m, nil
	}

	// Forward key to text input, then clear modal cache so it rebuilds
	var cmd tea.Cmd
	m.issueInputInput, cmd = m.issueInputInput.Update(msg)
	m.issueInputModal = nil
	m.issueInputModalWidth = 0

	// Trigger search if input changed (min 2 chars)
	newValue := strings.TrimSpace(m.issueInputInput.Value())
	if newValue != m.issueSearchQuery && len(newValue) >= 2 {
		m.issueSearchQuery = newValue
		m.issueSearchLoading = true
		// Keep previous results visible while loading to avoid modal shrink/grow flicker.
		// Results are replaced when the new IssueSearchResultMsg arrives.
		m.issueSearchCursor = -1
		return m, tea.Batch(cmd, issueSearchCmd(m.ui.WorkDir, newValue, m.issueSearchIncludeClosed))
	}
	if len(newValue) < 2 {
		m.issueSearchResults = nil
		m.issueSearchQuery = ""
		m.issueSearchCursor = -1
	}
	return m, cmd
}

// handleIssuePreviewKeys handles keyboard input for the issue preview modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleIssuePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.ensureIssuePreviewModal()
	if m.issuePreviewModal == nil {
		return m, nil
	}

	// Shortcuts before modal.HandleKey (which consumes Enter/Esc/Tab)
	switch msg.String() {
	case "j", "down":
		m.issuePreviewModal.ScrollBy(1)
		return m, nil
	case "k", "up":
		m.issuePreviewModal.ScrollBy(-1)
		return m, nil
	case "ctrl+d":
		m.issuePreviewModal.ScrollBy(10)
		return m, nil
	case "ctrl+u":
		m.issuePreviewModal.ScrollBy(-10)
		return m, nil
	case "g":
		m.issuePreviewModal.ScrollToTop()
		return m, nil
	case "G":
		m.issuePreviewModal.ScrollToBottom()
		return m, nil
	case "o":
		if m.issuePreviewData != nil {
			issueID := m.issuePreviewData.ID
			m.resetIssuePreview()
			m.resetIssueInput()
			m.updateContext()
			return m, tea.Batch(
				FocusPlugin("td-monitor"),
				func() tea.Msg { return OpenFullIssueMsg{IssueID: issueID} },
			)
		}
	case "b":
		m.backToIssueInput()
		return m, nil
	case "y":
		if m.issuePreviewData != nil {
			d := m.issuePreviewData
			text := d.ID + ": " + d.Title + "\n\n" + d.Description
			if err := clipboard.WriteAll(text); err != nil {
				return m, ShowToast("Copy failed: "+err.Error(), 2*time.Second)
			}
			return m, ShowToast("Yanked issue details", 2*time.Second)
		}
	case "Y":
		if m.issuePreviewData != nil {
			id := m.issuePreviewData.ID
			if err := clipboard.WriteAll(id); err != nil {
				return m, ShowToast("Copy failed: "+err.Error(), 2*time.Second)
			}
			return m, ShowToast("Yanked: "+id, 2*time.Second)
		}
	}

	action, cmd := m.issuePreviewModal.HandleKey(msg)
	switch action {
	case "open-in-td":
		issueID := ""
		if m.issuePreviewData != nil {
			issueID = m.issuePreviewData.ID
		}
		m.resetIssuePreview()
		m.resetIssueInput()
		m.updateContext()
		if issueID != "" {
			return m, tea.Batch(
				FocusPlugin("td-monitor"),
				func() tea.Msg { return OpenFullIssueMsg{IssueID: issueID} },
			)
		}
		return m, nil
	case "back":
		m.backToIssueInput()
		return m, nil
	case "cancel":
		m.resetIssuePreview()
		m.resetIssueInput()
		m.updateContext()
		return m, nil
	}
	return m, cmd
}

// issueInputSubmit resolves the current issue input (selected result or typed ID)
// and either opens the full issue in TD monitor or shows a lightweight preview.
func (m *Model) issueInputSubmit() (tea.Model, tea.Cmd) {
	var issueID string
	if m.issueSearchCursor >= 0 && m.issueSearchCursor < len(m.issueSearchResults) {
		issueID = m.issueSearchResults[m.issueSearchCursor].ID
	} else {
		issueID = strings.TrimSpace(m.issueInputInput.Value())
	}
	if issueID == "" {
		return m, nil
	}
	// Check if active plugin is TD monitor — go directly to rich modal
	if p := m.ActivePlugin(); p != nil && p.ID() == "td-monitor" {
		m.resetIssueInput()
		m.updateContext()
		return m, tea.Batch(
			func() tea.Msg { return OpenFullIssueMsg{IssueID: issueID} },
		)
	}
	// Hide input modal but preserve search state so "back" can restore it.
	m.showIssueInput = false
	// Show lightweight preview
	m.showIssuePreview = true
	m.activeContext = "issue-preview"
	m.issuePreviewLoading = true
	m.issuePreviewData = nil
	m.issuePreviewError = nil
	m.issuePreviewModal = nil
	m.issuePreviewModalWidth = 0
	m.issuePreviewMouseHandler = mouse.NewHandler()
	return m, fetchIssuePreviewCmd(m.ui.WorkDir, issueID)
}

// handleIssueInputMouse handles mouse events for the issue input modal.
func (m *Model) handleIssueInputMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureIssueInputModal()
	if m.issueInputModal == nil {
		return m, nil
	}
	if m.issueInputMouseHandler == nil {
		m.issueInputMouseHandler = mouse.NewHandler()
	}
	// Pre-render to sync hit regions and focusIDs on the (potentially rebuilt) modal.
	// The issue input modal is nilled on every keystroke to fix a stale text-input
	// pointer, so the modal object seen here may lack focusIDs from a prior Render.
	m.issueInputModal.Render(m.width, m.height, m.issueInputMouseHandler)
	action := m.issueInputModal.HandleMouse(msg, m.issueInputMouseHandler)
	switch {
	case action == "cancel":
		m.resetIssueInput()
		m.updateContext()
	case action == "open":
		return m.issueInputSubmit()
	case strings.HasPrefix(action, issueSearchResultPrefix):
		// Click on a search result — select it and submit
		idxStr := strings.TrimPrefix(action, issueSearchResultPrefix)
		if idx, err := strconv.Atoi(idxStr); err == nil && idx >= 0 && idx < len(m.issueSearchResults) {
			m.issueSearchCursor = idx
			return m.issueInputSubmit()
		}
	}
	return m, nil
}

// handleIssuePreviewMouse handles mouse events for the issue preview modal.
func (m *Model) handleIssuePreviewMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureIssuePreviewModal()
	if m.issuePreviewModal == nil {
		return m, nil
	}
	if m.issuePreviewMouseHandler == nil {
		m.issuePreviewMouseHandler = mouse.NewHandler()
	}
	// Pre-render to sync hit regions and focusIDs on the modal, which may have
	// been rebuilt (e.g. after data/error arrival cleared the cache).
	m.issuePreviewModal.Render(m.width, m.height, m.issuePreviewMouseHandler)
	action := m.issuePreviewModal.HandleMouse(msg, m.issuePreviewMouseHandler)
	switch action {
	case "cancel":
		m.resetIssuePreview()
		m.resetIssueInput()
		m.updateContext()
	case "back":
		m.backToIssueInput()
		return m, nil
	case "open-in-td":
		issueID := ""
		if m.issuePreviewData != nil {
			issueID = m.issuePreviewData.ID
		}
		m.resetIssuePreview()
		m.resetIssueInput()
		m.updateContext()
		if issueID != "" {
			return m, tea.Batch(
				FocusPlugin("td-monitor"),
				func() tea.Msg { return OpenFullIssueMsg{IssueID: issueID} },
			)
		}
	}
	return m, nil
}
