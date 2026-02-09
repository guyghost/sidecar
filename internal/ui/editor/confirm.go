package editor

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/styles"
)

// ExitChoice represents the user's selection in the exit confirmation dialog.
type ExitChoice int

const (
	ChoiceSaveAndExit ExitChoice = 0
	ChoiceExitNoSave  ExitChoice = 1
	ChoiceCancel      ExitChoice = 2
)

const exitChoiceCount = 3

// ExitConfirmation is a reusable exit confirmation dialog widget.
type ExitConfirmation struct {
	visible   bool
	selection int
}

// NewExitConfirmation creates a new exit confirmation widget.
func NewExitConfirmation() *ExitConfirmation {
	return &ExitConfirmation{}
}

// Show makes the confirmation dialog visible, resetting selection to SaveAndExit.
func (c *ExitConfirmation) Show() {
	c.visible = true
	c.selection = 0
}

// Hide hides the confirmation dialog.
func (c *ExitConfirmation) Hide() {
	c.visible = false
	c.selection = 0
}

// IsVisible returns whether the confirmation dialog is currently shown.
func (c *ExitConfirmation) IsVisible() bool {
	return c.visible
}

// Selection returns the current selection as an ExitChoice.
func (c *ExitConfirmation) Selection() ExitChoice {
	return ExitChoice(c.selection)
}

// HandleKey processes keyboard input for the confirmation dialog.
// Returns true if the key was handled, and an ExitChoice if Enter was pressed.
func (c *ExitConfirmation) HandleKey(msg tea.KeyMsg) (handled bool, chosen bool) {
	if !c.visible {
		return false, false
	}

	switch msg.String() {
	case "j", "down":
		c.selection = (c.selection + 1) % exitChoiceCount
		return true, false
	case "k", "up":
		c.selection = (c.selection - 1 + exitChoiceCount) % exitChoiceCount
		return true, false
	case "enter":
		return true, true
	case "esc":
		c.selection = int(ChoiceCancel)
		return true, true
	}

	return true, false // Absorb all other keys when visible
}

// Render renders the exit confirmation dialog.
func (c *ExitConfirmation) Render() string {
	options := []string{"Save & Exit", "Exit without saving", "Cancel"}

	var sb strings.Builder
	sb.WriteString(styles.Title.Render("Exit editor?"))
	sb.WriteString("\n\n")

	for i, opt := range options {
		if i == c.selection {
			sb.WriteString(styles.ListItemSelected.Render("> " + opt))
		} else {
			sb.WriteString("  " + opt)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(styles.Muted.Render("[j/k to select, Enter to confirm, Esc to cancel]"))

	return sb.String()
}
