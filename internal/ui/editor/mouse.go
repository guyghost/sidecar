package editor

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guyghost/sidecar/internal/tty"
)

// ForwardMousePress sends a mouse press event to the tmux session.
// col and row are 1-indexed coordinates relative to the editor content area.
func ForwardMousePress(sessionName string, col, row int) tea.Cmd {
	return func() tea.Msg {
		if err := tty.SendSGRMouse(sessionName, 0, col, row, false); err != nil {
			if tty.IsSessionDeadError(err) {
				return tty.SessionDeadMsg{}
			}
		}
		return nil
	}
}

// ForwardMouseDrag sends a mouse drag/motion event to the tmux session.
// col and row are 1-indexed coordinates relative to the editor content area.
func ForwardMouseDrag(sessionName string, col, row int) tea.Cmd {
	return func() tea.Msg {
		if err := tty.SendSGRMouse(sessionName, 32, col, row, false); err != nil {
			if tty.IsSessionDeadError(err) {
				return tty.SessionDeadMsg{}
			}
		}
		return nil
	}
}

// ForwardMouseRelease sends a mouse release event to the tmux session.
// col and row are 1-indexed coordinates relative to the editor content area.
func ForwardMouseRelease(sessionName string, col, row int) tea.Cmd {
	return func() tea.Msg {
		if err := tty.SendSGRMouse(sessionName, 0, col, row, true); err != nil {
			if tty.IsSessionDeadError(err) {
				return tty.SessionDeadMsg{}
			}
		}
		return nil
	}
}
