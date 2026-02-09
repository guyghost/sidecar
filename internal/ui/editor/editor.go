package editor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/guyghost/sidecar/internal/features"
)

// NormalizeEditorName normalizes editor names to canonical identifiers.
// Maps variations like nvim/neovim → vim, hx → helix, etc.
func NormalizeEditorName(editor string) string {
	base := filepath.Base(editor)
	base = strings.TrimSuffix(base, ".exe")
	switch base {
	case "nvim", "neovim":
		return "vim"
	case "vi":
		return "vim"
	case "hx":
		return "helix"
	case "kak":
		return "kakoune"
	case "emacsclient":
		return "emacs"
	}
	return base
}

// SendSaveAndQuit sends the key sequence to save and quit the editor.
// Returns true if the editor was recognized and keys were sent, false otherwise.
func SendSaveAndQuit(target, editor string) bool {
	normalized := NormalizeEditorName(editor)
	send := func(keys ...string) {
		for _, k := range keys {
			_ = exec.Command("tmux", "send-keys", "-t", target, k).Run()
		}
	}
	switch normalized {
	case "vim":
		send("Escape", ":wq", "Enter")
		return true
	case "nano":
		send("C-o", "Enter", "C-x")
		return true
	case "emacs":
		send("C-x", "C-s", "C-x", "C-c")
		return true
	case "helix":
		send("Escape", ":wq", "Enter")
		return true
	case "micro":
		send("C-s", "C-q")
		return true
	case "kakoune":
		send("Escape", ":write-quit", "Enter")
		return true
	case "joe":
		send("C-k", "x")
		return true
	case "ne":
		send("Escape", "Escape", ":s", "Enter", ":q", "Enter")
		return true
	case "amp":
		send("Escape", ":wq", "Enter")
		return true
	default:
		return false
	}
}

// SendCursorToEnd sends key sequence to move cursor to the end of the file.
func SendCursorToEnd(target, editor string) {
	normalized := NormalizeEditorName(editor)
	send := func(keys ...string) {
		for _, k := range keys {
			_ = exec.Command("tmux", "send-keys", "-t", target, k).Run()
		}
	}
	switch normalized {
	case "vim":
		send("G", "$")
	case "nano":
		send("M-/")
	case "emacs":
		send("M->")
	case "helix":
		send("g", "e")
	case "micro":
		send("C-End")
	case "kakoune":
		send("g", "e")
	}
}

// IsSessionAlive checks if a tmux session with the given name exists.
func IsSessionAlive(sessionName string) bool {
	if sessionName == "" {
		return false
	}
	err := exec.Command("tmux", "has-session", "-t", sessionName).Run()
	return err == nil
}

// KillSession kills the tmux session with the given name.
func KillSession(sessionName string) {
	if sessionName == "" {
		return
	}
	_ = exec.Command("tmux", "kill-session", "-t", sessionName).Run()
}

// IsSupported checks if inline editing is supported (feature flag + tmux available).
func IsSupported() bool {
	if !features.IsEnabled(features.TmuxInlineEdit.Name) {
		return false
	}
	if _, err := exec.LookPath("tmux"); err != nil {
		return false
	}
	return true
}

// ResolveEditor resolves the editor to use from environment variables.
// Falls back to EDITOR → VISUAL → "vim".
func ResolveEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	return "vim"
}

// ResolveTerm resolves the TERM environment variable.
// Falls back to "xterm-256color" if not set.
func ResolveTerm() string {
	if t := os.Getenv("TERM"); t != "" {
		return t
	}
	return "xterm-256color"
}
