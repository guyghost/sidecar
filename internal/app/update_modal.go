package app

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/guyghost/sidecar/internal/markdown"
	"github.com/guyghost/sidecar/internal/modal"
	"github.com/guyghost/sidecar/internal/mouse"
	"github.com/guyghost/sidecar/internal/styles"
	"github.com/guyghost/sidecar/internal/ui"
	"github.com/guyghost/sidecar/internal/version"
)

const changelogURL = "https://raw.githubusercontent.com/guyghost/sidecar/main/CHANGELOG.md"

// changelogViewState holds mutable state shared between the model and the
// modal's Custom section closure. Using a heap-allocated struct avoids
// rebuilding the modal on every scroll event (bubbletea value semantics
// would otherwise make the closure capture a stale model pointer).
type changelogViewState struct {
	ScrollOffset    int
	RenderedLines   []string
	MaxVisibleLines int
}

// updateModalWidth returns the appropriate modal width based on screen size.
func (m *Model) updateModalWidth() int {
	modalW := 60
	maxW := m.width - 4
	if maxW < 20 {
		maxW = 20 // Absolute minimum for very small screens
	}
	if modalW > maxW {
		modalW = maxW
	}
	if modalW < 30 {
		modalW = 30
	}
	// Final cap: never exceed available space
	if modalW > maxW {
		modalW = maxW
	}
	return modalW
}

// renderUpdateModalOverlay renders the update modal as an overlay on top of background.
func (m *Model) renderUpdateModalOverlay(background string) string {
	// Render modal content based on state
	var modalContent string
	switch m.update.ModalState {
	case UpdateModalPreview:
		modalContent = m.renderUpdatePreviewModal()
	case UpdateModalProgress:
		modalContent = m.renderUpdateProgressModal()
	case UpdateModalComplete:
		modalContent = m.renderUpdateCompleteModal()
	case UpdateModalError:
		modalContent = m.renderUpdateErrorModal()
	default:
		return background
	}

	return ui.OverlayModal(background, modalContent, m.width, m.height)
}

// ensureUpdatePreviewModal creates/updates the preview modal with caching.
func (m *Model) ensureUpdatePreviewModal() {
	if m.updateAvailable == nil && (m.tdVersionInfo == nil || !m.tdVersionInfo.HasUpdate) {
		return
	}
	modalW := m.updateModalWidth()
	if m.update.PreviewModal != nil && m.update.PreviewModalWidth == modalW {
		return
	}
	m.update.PreviewModalWidth = modalW
	contentW := modalW - 6 // borders + padding

	// TD-only update: build a simpler modal
	if m.updateAvailable == nil {
		arrow := lipgloss.NewStyle().Foreground(styles.Success).Render(" → ")
		versionLine := fmt.Sprintf("%s%s%s", m.tdVersionInfo.CurrentVersion, arrow, m.tdVersionInfo.LatestVersion)

		var methodHint string
		switch m.update.InstallMethod {
		case version.InstallMethodHomebrew:
			methodHint = styles.Muted.Render("Method: brew upgrade td")
		default:
			methodHint = styles.Muted.Render("Method: go install")
		}

		m.update.PreviewModal = modal.New("td Update",
			modal.WithWidth(modalW),
			modal.WithVariant(modal.VariantDefault),
			modal.WithPrimaryAction("update"),
		).
			AddSection(modal.Text(versionLine)).
			AddSection(modal.Spacer()).
			AddSection(modal.Text(methodHint)).
			AddSection(modal.Spacer()).
			AddSection(modal.Buttons(
				modal.Btn(" Update Now ", "update"),
				modal.Btn(" Later ", "cancel"),
			))
		return
	}

	// Sidecar update (possibly with td update bundled)
	// Version line
	currentV := m.updateAvailable.CurrentVersion
	latestV := m.updateAvailable.LatestVersion
	arrow := lipgloss.NewStyle().Foreground(styles.Success).Render(" → ")
	versionLine := fmt.Sprintf("%s%s%s", currentV, arrow, latestV)

	// Release notes
	releaseNotes := m.update.ReleaseNotes
	if releaseNotes == "" {
		releaseNotes = m.updateAvailable.ReleaseNotes
	}
	if releaseNotes == "" {
		releaseNotes = "No release notes available."
	}
	releaseNotes = parseReleaseNotes(releaseNotes)
	renderedNotes := m.renderReleaseNotes(releaseNotes, contentW)

	// Limit height
	lines := strings.Split(renderedNotes, "\n")
	maxLines := 15
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, styles.Muted.Render("... (truncated)"))
	}
	notesContent := strings.Join(lines, "\n")

	changelogHint := styles.Muted.Render("[c] View Full Changelog")

	// Build method-specific install hint and buttons
	var methodHint string
	var buttons []modal.ButtonDef

	switch m.update.InstallMethod {
	case version.InstallMethodHomebrew:
		methodHint = styles.Muted.Render("Method: brew upgrade sidecar")
		buttons = []modal.ButtonDef{
			modal.Btn(" Update Now ", "update"),
			modal.Btn(" Later ", "cancel"),
		}
	case version.InstallMethodBinary:
		downloadURL := fmt.Sprintf("https://github.com/guyghost/sidecar/releases/tag/%s",
			m.updateAvailable.LatestVersion)
		methodHint = styles.Muted.Render("Download: " + downloadURL)
		buttons = []modal.ButtonDef{
			modal.Btn(" Close ", "cancel"),
		}
	default:
		methodHint = styles.Muted.Render("Method: go install")
		buttons = []modal.ButtonDef{
			modal.Btn(" Update Now ", "update"),
			modal.Btn(" Later ", "cancel"),
		}
	}

	m.update.PreviewModal = modal.New("Sidecar Update",
		modal.WithWidth(modalW),
		modal.WithVariant(modal.VariantDefault),
		modal.WithPrimaryAction("update"),
	).
		AddSection(modal.Text(versionLine)).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(lipgloss.NewStyle().Bold(true).Render("What's New"))).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(notesContent)).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(changelogHint)).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(methodHint)).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(buttons...))
}

// renderUpdatePreviewModal renders the preview state showing release notes before update.
func (m *Model) renderUpdatePreviewModal() string {
	m.ensureUpdatePreviewModal()
	if m.update.PreviewModal == nil {
		return ""
	}
	if m.update.PreviewMouseHandler == nil {
		m.update.PreviewMouseHandler = mouse.NewHandler()
	}
	return m.update.PreviewModal.Render(m.width, m.height, m.update.PreviewMouseHandler)
}

// parseReleaseNotes cleans up release notes by removing duplicate headers
// and excessive whitespace. The modal already shows "What's New" as a header,
// so we strip any leading "What's New" headers from the content.
func parseReleaseNotes(notes string) string {
	if notes == "" {
		return notes
	}

	// Patterns to strip from the beginning of release notes
	// Match: ## What's New, ### What's New, # What's New (case-insensitive)
	// Also match: # Release Notes, ## Release Notes
	headerPatterns := regexp.MustCompile(`(?im)^#+\s*(what'?s?\s*new|release\s*notes)\s*\n*`)

	result := notes

	// Strip leading whitespace and newlines first
	result = strings.TrimSpace(result)

	// Repeatedly strip matching headers from the beginning
	// (in case there are multiple, e.g., "## What's New\n### What's New\n")
	for {
		loc := headerPatterns.FindStringIndex(result)
		if loc == nil || loc[0] != 0 {
			break
		}
		result = result[loc[1]:]
		result = strings.TrimSpace(result)
	}

	// Collapse multiple consecutive newlines to at most 2
	multiNewlines := regexp.MustCompile(`\n{3,}`)
	result = multiNewlines.ReplaceAllString(result, "\n\n")

	// If parsing emptied the content, return original
	if strings.TrimSpace(result) == "" {
		return strings.TrimSpace(notes)
	}

	return result
}

// renderReleaseNotes renders markdown release notes.
func (m *Model) renderReleaseNotes(notes string, width int) string {
	// Try to use markdown renderer
	renderer, err := markdown.NewRenderer()
	if err != nil {
		return notes
	}

	lines := renderer.RenderContent(notes, width)
	return strings.Join(lines, "\n")
}

// centerText centers text within a given width.
func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

// renderUpdateProgressModal renders the progress state during update.
func (m *Model) renderUpdateProgressModal() string {
	modalW := m.updateModalWidth()
	contentW := modalW - 4

	var sb strings.Builder

	// Title
	title := lipgloss.NewStyle().Bold(true).Foreground(styles.Warning).Render("Updating Sidecar")
	sb.WriteString(centerText(title, contentW))
	sb.WriteString("\n\n")

	// Version being installed
	if m.updateAvailable != nil {
		version := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(
			fmt.Sprintf("Installing %s", m.updateAvailable.LatestVersion))
		sb.WriteString(centerText(version, contentW))
		sb.WriteString("\n\n")
	}

	// Phase indicators - 3 real, observable phases
	phases := []UpdatePhase{PhaseCheckPrereqs, PhaseInstalling, PhaseVerifying}
	methodStr := string(m.update.InstallMethod)
	for _, phase := range phases {
		status := m.update.PhaseStatus[phase]
		icon := "○" // pending
		color := styles.TextMuted

		switch status {
		case "running":
			icon = "●"
			color = styles.Warning
		case "done":
			icon = "✓"
			color = styles.Success
		case "error":
			icon = "✗"
			color = styles.Error
		}

		phaseName := phase.StringForMethod(methodStr)
		if phase == m.update.Phase && status == "running" {
			phaseName = lipgloss.NewStyle().Bold(true).Render(phaseName)
		}

		phaseLine := fmt.Sprintf("  %s %s",
			lipgloss.NewStyle().Foreground(color).Render(icon),
			phaseName,
		)
		sb.WriteString(phaseLine)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Elapsed time
	elapsed := m.getUpdateElapsed()
	elapsedStr := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(
		fmt.Sprintf("Elapsed: %s", formatElapsed(elapsed)))
	sb.WriteString(centerText(elapsedStr, contentW))
	sb.WriteString("\n\n")

	// Divider
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.TextMuted).Render(strings.Repeat("─", contentW)))
	sb.WriteString("\n\n")

	// Cancel hint
	cancelHint := lipgloss.NewStyle().Foreground(styles.TextMuted).Render("Esc: cancel")
	sb.WriteString(centerText(cancelHint, contentW))

	// Constrain modal height to available space per CLAUDE.md
	maxHeight := m.height - 4
	if maxHeight < 10 {
		maxHeight = 10
	}

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.TextMuted).
		Padding(1, 2).
		Width(modalW).
		MaxHeight(maxHeight)

	return modalStyle.Render(sb.String())
}

// getUpdateElapsed returns the elapsed time since update started.
func (m *Model) getUpdateElapsed() time.Duration {
	if m.update.StartTime.IsZero() {
		return 0
	}
	return time.Since(m.update.StartTime)
}

// formatElapsed formats a duration as M:SS.
func formatElapsed(d time.Duration) string {
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// ensureUpdateCompleteModal creates/updates the complete modal with caching.
func (m *Model) ensureUpdateCompleteModal() {
	modalW := m.updateModalWidth()
	if m.update.CompleteModal != nil && m.update.CompleteModalWidth == modalW {
		return
	}
	m.update.CompleteModalWidth = modalW

	checkmark := lipgloss.NewStyle().Foreground(styles.Success).Render("✓")

	var updatesText string
	if m.updateAvailable != nil {
		updatesText = fmt.Sprintf("  %s Sidecar updated to %s", checkmark, m.updateAvailable.LatestVersion)
	} else {
		updatesText = fmt.Sprintf("  %s Sidecar updated", checkmark)
	}
	if m.tdVersionInfo != nil && m.tdVersionInfo.HasUpdate {
		updatesText += fmt.Sprintf("\n  %s td updated to %s", checkmark, m.tdVersionInfo.LatestVersion)
	}

	restartMsg := styles.Muted.Render("Restart sidecar to use the new version.")
	tip := styles.Muted.Render("Tip: Press q to quit, then run 'sidecar' again.")

	m.update.CompleteModal = modal.New("Update Complete!",
		modal.WithWidth(modalW),
		modal.WithVariant(modal.VariantInfo),
		modal.WithPrimaryAction("quit"),
	).
		AddSection(modal.Text(updatesText)).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(restartMsg)).
		AddSection(modal.Text(tip)).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(
			modal.Btn(" Quit & Restart ", "quit"),
			modal.Btn(" Later ", "cancel"),
		))
}

// renderUpdateCompleteModal renders the completion state.
func (m *Model) renderUpdateCompleteModal() string {
	m.ensureUpdateCompleteModal()
	if m.update.CompleteModal == nil {
		return ""
	}
	if m.update.CompleteMouseHandler == nil {
		m.update.CompleteMouseHandler = mouse.NewHandler()
	}
	return m.update.CompleteModal.Render(m.width, m.height, m.update.CompleteMouseHandler)
}

// ensureUpdateErrorModal creates/updates the error modal with caching.
func (m *Model) ensureUpdateErrorModal() {
	modalW := m.updateModalWidth()
	if m.update.ErrorModal != nil && m.update.ErrorModalWidth == modalW {
		return
	}
	m.update.ErrorModalWidth = modalW

	errorIcon := lipgloss.NewStyle().Foreground(styles.Error).Render("✗")
	phaseName := m.update.Phase.String()
	errorLine := fmt.Sprintf("  %s Error during: %s", errorIcon, phaseName)

	errorMsg := m.update.Error
	if errorMsg == "" {
		errorMsg = "An unknown error occurred."
	}

	m.update.ErrorModal = modal.New("Update Failed",
		modal.WithWidth(modalW),
		modal.WithVariant(modal.VariantDanger),
		modal.WithPrimaryAction("retry"),
	).
		AddSection(modal.Text(errorLine)).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(styles.Muted.Render("  " + errorMsg))).
		AddSection(modal.Spacer()).
		AddSection(modal.Buttons(
			modal.Btn(" Retry ", "retry"),
			modal.Btn(" Close ", "cancel"),
		))
}

// renderUpdateErrorModal renders the error state.
func (m *Model) renderUpdateErrorModal() string {
	m.ensureUpdateErrorModal()
	if m.update.ErrorModal == nil {
		return ""
	}
	if m.update.ErrorMouseHandler == nil {
		m.update.ErrorMouseHandler = mouse.NewHandler()
	}
	return m.update.ErrorModal.Render(m.width, m.height, m.update.ErrorMouseHandler)
}

// getChangelogModalWidth returns the width for the changelog modal.
func (m *Model) getChangelogModalWidth() int {
	modalW := m.updateModalWidth() + 10
	maxW := m.width - 4
	if modalW > maxW {
		modalW = maxW
	}
	if modalW < 30 {
		modalW = 30
	}
	return modalW
}

// ensureChangelogModal creates/updates the changelog modal with caching.
// The modal is only rebuilt when width or height changes. Scroll offset changes
// are handled dynamically via the shared changelogScrollState pointer.
func (m *Model) ensureChangelogModal() {
	modalW := m.getChangelogModalWidth()
	contentW := modalW - 6 // borders + padding

	// Calculate max visible lines
	modalMaxHeight := m.height - 6
	if modalMaxHeight < 10 {
		modalMaxHeight = 10
	}
	maxContentLines := modalMaxHeight - 8
	if maxContentLines < 5 {
		maxContentLines = 5
	}

	// Check if we can reuse the cached modal
	// Rebuild only if width or max visible lines changed
	if m.update.ChangelogModal != nil &&
		m.update.ChangelogModalWidth == modalW &&
		m.update.ChangelogMaxVisibleLines == maxContentLines {
		return
	}

	m.update.ChangelogModalWidth = modalW
	m.update.ChangelogMaxVisibleLines = maxContentLines

	// Render changelog content and cache the lines
	content := m.update.Changelog
	if content == "" {
		content = "Loading changelog..."
	}
	renderedContent := m.renderReleaseNotes(content, contentW)
	m.update.ChangelogRenderedLines = strings.Split(renderedContent, "\n")

	// Initialize or update the shared scroll state
	if m.update.ChangelogScrollState == nil {
		m.update.ChangelogScrollState = &changelogViewState{}
	}
	m.update.ChangelogScrollState.RenderedLines = m.update.ChangelogRenderedLines
	m.update.ChangelogScrollState.MaxVisibleLines = maxContentLines

	// Capture shared pointer - survives bubbletea value copies
	state := m.update.ChangelogScrollState

	// Create a custom section that handles scrolling dynamically.
	// The closure reads from the shared state pointer so scroll changes
	// don't require rebuilding the modal.
	scrollSection := modal.Custom(func(cw int, focusID, hoverID string) modal.RenderedSection {
		lines := state.RenderedLines
		maxVisible := state.MaxVisibleLines

		// Apply scroll offset with clamping
		startLine := state.ScrollOffset
		maxStart := len(lines) - maxVisible
		if maxStart < 0 {
			maxStart = 0
		}
		if startLine > maxStart {
			startLine = maxStart
		}
		if startLine < 0 {
			startLine = 0
		}

		endLine := startLine + maxVisible
		if endLine > len(lines) {
			endLine = len(lines)
		}

		visibleLines := lines[startLine:endLine]
		visibleContent := strings.Join(visibleLines, "\n")

		// Add scroll indicator if needed
		if len(lines) > maxVisible {
			scrollInfo := styles.Muted.Render(fmt.Sprintf("Lines %d-%d of %d", startLine+1, endLine, len(lines)))
			visibleContent += "\n\n" + scrollInfo
		}

		return modal.RenderedSection{Content: visibleContent}
	}, nil)

	m.update.ChangelogModal = modal.New("Changelog",
		modal.WithWidth(modalW),
		modal.WithVariant(modal.VariantDefault),
		modal.WithHints(false), // We show custom hints
	).
		AddSection(scrollSection).
		AddSection(modal.Spacer()).
		AddSection(modal.Text(styles.Muted.Render("j/k scroll   Esc: close"))).
		AddSection(modal.Buttons(
			modal.Btn(" Close ", "cancel"),
		))
}

// clearChangelogModal clears the changelog modal cache.
func (m *Model) clearChangelogModal() {
	m.update.ChangelogModal = nil
	m.update.ChangelogModalWidth = 0
	m.update.ChangelogMouseHandler = nil
	m.update.ChangelogRenderedLines = nil
	m.update.ChangelogMaxVisibleLines = 0
	m.update.ChangelogScrollState = nil
}

// syncChangelogScroll updates the shared scroll state from the model field.
// Call this after modifying changelogScrollOffset instead of clearChangelogModal.
func (m *Model) syncChangelogScroll() {
	if m.update.ChangelogScrollState != nil {
		m.update.ChangelogScrollState.ScrollOffset = m.update.ChangelogScrollOffset
	}
}

// fetchChangelog fetches the CHANGELOG.md from GitHub.
func fetchChangelog() tea.Cmd {
	return func() tea.Msg {
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(changelogURL)
		if err != nil {
			return ChangelogLoadedMsg{Err: err}
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return ChangelogLoadedMsg{Err: fmt.Errorf("HTTP %d", resp.StatusCode)}
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return ChangelogLoadedMsg{Err: err}
		}

		return ChangelogLoadedMsg{Content: string(body)}
	}
}

// renderChangelogOverlay renders the changelog as an overlay on the update preview modal.
func (m *Model) renderChangelogOverlay(background string) string {
	m.ensureChangelogModal()
	if m.update.ChangelogModal == nil {
		return background
	}
	if m.update.ChangelogMouseHandler == nil {
		m.update.ChangelogMouseHandler = mouse.NewHandler()
	}
	modalContent := m.update.ChangelogModal.Render(m.width, m.height, m.update.ChangelogMouseHandler)
	return ui.OverlayModal(background, modalContent, m.width, m.height)
}
