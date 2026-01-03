// Package ui provides shared UI components and helpers for the TUI.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// DimStyle is used to dim background content behind modals.
var DimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

// OverlayModal composites a modal on top of a dimmed background.
// The modal is centered, with dimmed background visible above and below.
func OverlayModal(background, modal string, width, height int) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	// Dim background lines
	for i, line := range bgLines {
		stripped := ansi.Strip(line)
		bgLines[i] = DimStyle.Render(stripped)
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

	// Modal lines (centered horizontally)
	for _, line := range modalLines {
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
