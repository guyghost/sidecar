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
		m.issue.SearchIncludeClosed = !m.issue.SearchIncludeClosed
		m.issue.SearchScrollOffset = 0
		m.issue.SearchCursor = -1
		m.issue.InputModal = nil
		m.issue.InputModalWidth = 0
		if len(strings.TrimSpace(m.issue.InputModel.Value())) >= 2 {
			m.issue.SearchLoading = true
			return m, issueSearchCmd(m.ui.WorkDir, strings.TrimSpace(m.issue.InputModel.Value()), m.issue.SearchIncludeClosed)
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEnter:
		return m.issueInputSubmit()
	case tea.KeyUp:
		if len(m.issue.SearchResults) > 0 {
			m.issue.SearchCursor--
			if m.issue.SearchCursor < -1 {
				m.issue.SearchCursor = -1
			}
			// Keep cursor visible in viewport
			if m.issue.SearchCursor >= 0 && m.issue.SearchCursor < m.issue.SearchScrollOffset {
				m.issue.SearchScrollOffset = m.issue.SearchCursor
			}
			m.issue.InputModal = nil
			m.issue.InputModalWidth = 0
			return m, nil
		}
	case tea.KeyDown:
		if len(m.issue.SearchResults) > 0 {
			m.issue.SearchCursor++
			if m.issue.SearchCursor >= len(m.issue.SearchResults) {
				m.issue.SearchCursor = len(m.issue.SearchResults) - 1
			}
			// Keep cursor visible in viewport
			const maxVisible = 10
			if m.issue.SearchCursor >= m.issue.SearchScrollOffset+maxVisible {
				m.issue.SearchScrollOffset = m.issue.SearchCursor - maxVisible + 1
			}
			m.issue.InputModal = nil
			m.issue.InputModalWidth = 0
			return m, nil
		}
	case tea.KeyTab:
		if m.issue.SearchCursor >= 0 && m.issue.SearchCursor < len(m.issue.SearchResults) {
			m.issue.InputModel.SetValue(m.issue.SearchResults[m.issue.SearchCursor].ID)
			m.issue.InputModel.CursorEnd()
			m.issue.InputModal = nil
			m.issue.InputModalWidth = 0
		}
		// Tab is consumed (fill-in or no-op) — don't forward to textinput
		return m, nil
	}

	if isMouseEscapeSequence(msg) {
		return m, nil
	}

	// Forward key to text input, then clear modal cache so it rebuilds
	var cmd tea.Cmd
	m.issue.InputModel, cmd = m.issue.InputModel.Update(msg)
	m.issue.InputModal = nil
	m.issue.InputModalWidth = 0

	// Trigger search if input changed (min 2 chars)
	newValue := strings.TrimSpace(m.issue.InputModel.Value())
	if newValue != m.issue.SearchQuery && len(newValue) >= 2 {
		m.issue.SearchQuery = newValue
		m.issue.SearchLoading = true
		// Keep previous results visible while loading to avoid modal shrink/grow flicker.
		// Results are replaced when the new IssueSearchResultMsg arrives.
		m.issue.SearchCursor = -1
		return m, tea.Batch(cmd, issueSearchCmd(m.ui.WorkDir, newValue, m.issue.SearchIncludeClosed))
	}
	if len(newValue) < 2 {
		m.issue.SearchResults = nil
		m.issue.SearchQuery = ""
		m.issue.SearchCursor = -1
	}
	return m, cmd
}

// handleIssuePreviewKeys handles keyboard input for the issue preview modal.
// Extracted from handleKeyMsg to reduce update.go complexity.
func (m *Model) handleIssuePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.ensureIssuePreviewModal()
	if m.issue.PreviewModal == nil {
		return m, nil
	}

	// Shortcuts before modal.HandleKey (which consumes Enter/Esc/Tab)
	switch msg.String() {
	case "j", "down":
		m.issue.PreviewModal.ScrollBy(1)
		return m, nil
	case "k", "up":
		m.issue.PreviewModal.ScrollBy(-1)
		return m, nil
	case "ctrl+d":
		m.issue.PreviewModal.ScrollBy(10)
		return m, nil
	case "ctrl+u":
		m.issue.PreviewModal.ScrollBy(-10)
		return m, nil
	case "g":
		m.issue.PreviewModal.ScrollToTop()
		return m, nil
	case "G":
		m.issue.PreviewModal.ScrollToBottom()
		return m, nil
	case "o":
		if m.issue.PreviewData != nil {
			issueID := m.issue.PreviewData.ID
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
		if m.issue.PreviewData != nil {
			d := m.issue.PreviewData
			text := d.ID + ": " + d.Title + "\n\n" + d.Description
			if err := clipboard.WriteAll(text); err != nil {
				return m, ShowToast("Copy failed: "+err.Error(), 2*time.Second)
			}
			return m, ShowToast("Yanked issue details", 2*time.Second)
		}
	case "Y":
		if m.issue.PreviewData != nil {
			id := m.issue.PreviewData.ID
			if err := clipboard.WriteAll(id); err != nil {
				return m, ShowToast("Copy failed: "+err.Error(), 2*time.Second)
			}
			return m, ShowToast("Yanked: "+id, 2*time.Second)
		}
	}

	action, cmd := m.issue.PreviewModal.HandleKey(msg)
	switch action {
	case "open-in-td":
		issueID := ""
		if m.issue.PreviewData != nil {
			issueID = m.issue.PreviewData.ID
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
	if m.issue.SearchCursor >= 0 && m.issue.SearchCursor < len(m.issue.SearchResults) {
		issueID = m.issue.SearchResults[m.issue.SearchCursor].ID
	} else {
		issueID = strings.TrimSpace(m.issue.InputModel.Value())
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
	m.issue.ShowInput = false
	// Show lightweight preview
	m.issue.ShowPreview = true
	m.activeContext = "issue-preview"
	m.issue.PreviewLoading = true
	m.issue.PreviewData = nil
	m.issue.PreviewError = nil
	m.issue.PreviewModal = nil
	m.issue.PreviewModalWidth = 0
	m.issue.PreviewMouseHandler = mouse.NewHandler()
	return m, fetchIssuePreviewCmd(m.ui.WorkDir, issueID)
}

// handleIssueInputMouse handles mouse events for the issue input modal.
func (m *Model) handleIssueInputMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureIssueInputModal()
	if m.issue.InputModal == nil {
		return m, nil
	}
	if m.issue.InputMouseHandler == nil {
		m.issue.InputMouseHandler = mouse.NewHandler()
	}
	// Pre-render to sync hit regions and focusIDs on the (potentially rebuilt) modal.
	// The issue input modal is nilled on every keystroke to fix a stale text-input
	// pointer, so the modal object seen here may lack focusIDs from a prior Render.
	m.issue.InputModal.Render(m.width, m.height, m.issue.InputMouseHandler)
	action := m.issue.InputModal.HandleMouse(msg, m.issue.InputMouseHandler)
	switch {
	case action == "cancel":
		m.resetIssueInput()
		m.updateContext()
	case action == "open":
		return m.issueInputSubmit()
	case strings.HasPrefix(action, issueSearchResultPrefix):
		// Click on a search result — select it and submit
		idxStr := strings.TrimPrefix(action, issueSearchResultPrefix)
		if idx, err := strconv.Atoi(idxStr); err == nil && idx >= 0 && idx < len(m.issue.SearchResults) {
			m.issue.SearchCursor = idx
			return m.issueInputSubmit()
		}
	}
	return m, nil
}

// handleIssuePreviewMouse handles mouse events for the issue preview modal.
func (m *Model) handleIssuePreviewMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.ensureIssuePreviewModal()
	if m.issue.PreviewModal == nil {
		return m, nil
	}
	if m.issue.PreviewMouseHandler == nil {
		m.issue.PreviewMouseHandler = mouse.NewHandler()
	}
	// Pre-render to sync hit regions and focusIDs on the modal, which may have
	// been rebuilt (e.g. after data/error arrival cleared the cache).
	m.issue.PreviewModal.Render(m.width, m.height, m.issue.PreviewMouseHandler)
	action := m.issue.PreviewModal.HandleMouse(msg, m.issue.PreviewMouseHandler)
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
		if m.issue.PreviewData != nil {
			issueID = m.issue.PreviewData.ID
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
