package gitstatus

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/marcus/sidecar/internal/styles"
)

// dimStyle is used to dim background content behind modals.
var dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// renderPushMenu renders the push options popup menu.
func (p *Plugin) renderPushMenu() string {
	// Render the background (current view dimmed)
	var background string
	switch p.pushMenuReturnMode {
	case ViewModeHistory:
		background = p.renderHistory()
	case ViewModeCommitDetail:
		background = p.renderCommitDetail()
	default:
		background = p.renderThreePaneView()
	}

	// Build menu content
	var sb strings.Builder

	// Menu options
	options := []struct{ key, label string }{
		{"p", "Push to origin"},
		{"f", "Force push (--force-with-lease)"},
		{"u", "Push & set upstream (-u)"},
	}

	for i, opt := range options {
		key := styles.KeyHint.Render(" " + opt.key + " ")
		sb.WriteString(fmt.Sprintf("  %s  %s", key, opt.label))
		if i < len(options)-1 {
			sb.WriteString("\n\n") // Spacing between options
		} else {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(styles.Muted.Render("  Esc to cancel"))

	// Create menu box - wide enough for longest option
	menuWidth := 44
	title := styles.Title.Render(" Push ")

	menuContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		Width(menuWidth).
		Render(title + "\n\n" + sb.String())

	// Overlay menu on dimmed background (overlayModal handles centering)
	return overlayModal(background, menuContent, p.width, p.height)
}

// overlayModal composites a modal on top of a dimmed background.
// The modal is centered, with dimmed background visible above and below.
func overlayModal(background, modal string, width, height int) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	// Dim background lines
	for i, line := range bgLines {
		stripped := ansi.Strip(line)
		bgLines[i] = dimStyle.Render(stripped)
	}

	// Ensure we have enough bg lines
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}

	// Calculate vertical center position
	modalHeight := len(modalLines)
	startY := (height - modalHeight) / 2
	if startY < 0 {
		startY = 0
	}

	// Build result: dimmed bg above, modal lines, dimmed bg below
	result := make([]string, 0, height)

	// Above modal: dimmed background
	for y := 0; y < startY && y < len(bgLines); y++ {
		result = append(result, bgLines[y])
	}

	// Modal lines (centered horizontally using lipgloss)
	for _, line := range modalLines {
		// Center each modal line
		lineWidth := ansi.StringWidth(line)
		leftPad := (width - lineWidth) / 2
		if leftPad < 0 {
			leftPad = 0
		}
		result = append(result, strings.Repeat(" ", leftPad)+line)
	}

	// Below modal: dimmed background
	for y := startY + modalHeight; y < height && y < len(bgLines); y++ {
		result = append(result, bgLines[y])
	}

	return strings.Join(result, "\n")
}
