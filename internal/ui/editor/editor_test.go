package editor

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNormalizeEditorName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// vim family
		{"vim", "vim", "vim"},
		{"nvim", "nvim", "vim"},
		{"neovim", "neovim", "vim"},
		{"vi", "vi", "vim"},

		// helix
		{"hx", "hx", "helix"},

		// kakoune
		{"kak", "kak", "kakoune"},

		// emacs
		{"emacsclient", "emacsclient", "emacs"},

		// Path handling
		{"vim path", "/usr/bin/vim", "vim"},
		{"nvim path", "/usr/local/bin/nvim", "vim"},

		// Passthrough (unchanged)
		{"nano", "nano", "nano"},
		{"emacs", "emacs", "emacs"},
		{"helix", "helix", "helix"},
		{"micro", "micro", "micro"},
		{"code", "code", "code"},

		// Windows suffix
		{"vim.exe", "vim.exe", "vim"},
		{"nvim.exe", "nvim.exe", "vim"},

		// Edge cases
		{"empty", "", "."},
		{"path with helix", "/path/to/helix", "helix"},
		{"nested path", "/usr/local/bin/kak", "kakoune"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeEditorName(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeEditorName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestResolveEditor(t *testing.T) {
	tests := []struct {
		name     string
		editor   string
		visual   string
		expected string
	}{
		{
			name:     "EDITOR set",
			editor:   "nano",
			visual:   "vim",
			expected: "nano",
		},
		{
			name:     "EDITOR empty, VISUAL set",
			editor:   "",
			visual:   "emacs",
			expected: "emacs",
		},
		{
			name:     "both set, EDITOR takes precedence",
			editor:   "nvim",
			visual:   "emacs",
			expected: "nvim",
		},
		{
			name:     "both empty, fallback to vim",
			editor:   "",
			visual:   "",
			expected: "vim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origEditor := os.Getenv("EDITOR")
			origVisual := os.Getenv("VISUAL")

			// Set test values
			if tt.editor != "" {
				os.Setenv("EDITOR", tt.editor)
			} else {
				os.Unsetenv("EDITOR")
			}
			if tt.visual != "" {
				os.Setenv("VISUAL", tt.visual)
			} else {
				os.Unsetenv("VISUAL")
			}

			// Test
			result := ResolveEditor()
			if result != tt.expected {
				t.Errorf("ResolveEditor() = %q, want %q", result, tt.expected)
			}

			// Restore original values
			if origEditor != "" {
				os.Setenv("EDITOR", origEditor)
			} else {
				os.Unsetenv("EDITOR")
			}
			if origVisual != "" {
				os.Setenv("VISUAL", origVisual)
			} else {
				os.Unsetenv("VISUAL")
			}
		})
	}
}

func TestResolveTerm(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		expected string
	}{
		{
			name:     "TERM set",
			term:     "screen-256color",
			expected: "screen-256color",
		},
		{
			name:     "TERM empty, fallback to xterm-256color",
			term:     "",
			expected: "xterm-256color",
		},
		{
			name:     "TERM is alacritty",
			term:     "alacritty",
			expected: "alacritty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			origTerm := os.Getenv("TERM")

			// Set test value
			if tt.term != "" {
				os.Setenv("TERM", tt.term)
			} else {
				os.Unsetenv("TERM")
			}

			// Test
			result := ResolveTerm()
			if result != tt.expected {
				t.Errorf("ResolveTerm() = %q, want %q", result, tt.expected)
			}

			// Restore original value
			if origTerm != "" {
				os.Setenv("TERM", origTerm)
			} else {
				os.Unsetenv("TERM")
			}
		})
	}
}

func TestExitConfirmation(t *testing.T) {
	t.Run("NewExitConfirmation", func(t *testing.T) {
		c := NewExitConfirmation()
		if c == nil {
			t.Fatal("NewExitConfirmation() returned nil")
		}
		if c.visible {
			t.Error("NewExitConfirmation() should start hidden")
		}
		if c.selection != 0 {
			t.Errorf("NewExitConfirmation() should start with selection 0, got %d", c.selection)
		}
	})

	t.Run("Show and IsVisible", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()
		if !c.visible {
			t.Error("Show() should make dialog visible")
		}
		if c.selection != 0 {
			t.Errorf("Show() should reset selection to 0, got %d", c.selection)
		}
		if !c.IsVisible() {
			t.Error("IsVisible() should return true after Show()")
		}
	})

	t.Run("Hide", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()
		c.Hide()
		if c.visible {
			t.Error("Hide() should make dialog hidden")
		}
		if c.selection != 0 {
			t.Errorf("Hide() should reset selection to 0, got %d", c.selection)
		}
		if c.IsVisible() {
			t.Error("IsVisible() should return false after Hide()")
		}
	})

	t.Run("Selection", func(t *testing.T) {
		c := NewExitConfirmation()

		c.selection = 0
		if c.Selection() != ChoiceSaveAndExit {
			t.Errorf("Selection() should return ChoiceSaveAndExit for selection 0")
		}

		c.selection = 1
		if c.Selection() != ChoiceExitNoSave {
			t.Errorf("Selection() should return ChoiceExitNoSave for selection 1")
		}

		c.selection = 2
		if c.Selection() != ChoiceCancel {
			t.Errorf("Selection() should return ChoiceCancel for selection 2")
		}
	})

	t.Run("HandleKey navigation", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()

		// Test 'j' - move down
		handled, chosen := c.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if !handled {
			t.Error("HandleKey should return handled=true for 'j'")
		}
		if chosen {
			t.Error("HandleKey should return chosen=false for 'j'")
		}
		if c.selection != 1 {
			t.Errorf("'j' should move selection from 0 to 1, got %d", c.selection)
		}

		// Test 'k' - move up
		handled, chosen = c.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if !handled {
			t.Error("HandleKey should return handled=true for 'k'")
		}
		if chosen {
			t.Error("HandleKey should return chosen=false for 'k'")
		}
		if c.selection != 0 {
			t.Errorf("'k' should move selection from 1 back to 0, got %d", c.selection)
		}

		// Test arrow keys
		handled, chosen = c.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
		if !handled {
			t.Error("HandleKey should return handled=true for down arrow")
		}
		if c.selection != 1 {
			t.Error("Down arrow should move selection to 1")
		}

		handled, chosen = c.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
		if !handled {
			t.Error("HandleKey should return handled=true for up arrow")
		}
		if c.selection != 0 {
			t.Error("Up arrow should move selection back to 0")
		}
	})

	t.Run("Selection wrapping", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()

		// Wrap from bottom to top
		c.selection = 2 // ChoiceCancel
		c.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if c.selection != 0 {
			t.Errorf("'j' from choice 2 should wrap to 0, got %d", c.selection)
		}

		// Wrap from top to bottom
		c.selection = 0 // ChoiceSaveAndExit
		c.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if c.selection != 2 {
			t.Errorf("'k' from choice 0 should wrap to 2, got %d", c.selection)
		}
	})

	t.Run("HandleKey enter", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()
		c.selection = 1 // ChoiceExitNoSave

		handled, chosen := c.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
		if !handled {
			t.Error("HandleKey should return handled=true for Enter")
		}
		if !chosen {
			t.Error("HandleKey should return chosen=true for Enter")
		}
		// Selection should be unchanged
		if c.selection != 1 {
			t.Errorf("Enter should not change selection, got %d", c.selection)
		}
	})

	t.Run("HandleKey esc", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()
		c.selection = 0 // ChoiceSaveAndExit

		handled, chosen := c.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
		if !handled {
			t.Error("HandleKey should return handled=true for Esc")
		}
		if !chosen {
			t.Error("HandleKey should return chosen=true for Esc")
		}
		// Selection should be changed to Cancel
		if c.selection != int(ChoiceCancel) {
			t.Errorf("Esc should change selection to ChoiceCancel, got %d", c.selection)
		}
	})

	t.Run("HandleKey absorbs all keys when visible", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()

		keys := []tea.KeyMsg{
			{Type: tea.KeyRunes, Runes: []rune{'x'}},
			{Type: tea.KeyTab},
			{Type: tea.KeySpace},
			{Type: tea.KeyBackspace},
			{Type: tea.KeyCtrlC},
		}

		for _, key := range keys {
			handled, chosen := c.HandleKey(key)
			if !handled {
				t.Errorf("HandleKey should absorb all keys when visible, got handled=false for %v", key)
			}
			if chosen {
				t.Errorf("HandleKey should return chosen=false for non-enter/esc keys, got chosen=true for %v", key)
			}
		}
	})

	t.Run("HandleKey when hidden", func(t *testing.T) {
		c := NewExitConfirmation()
		// Don't show it

		handled, chosen := c.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if handled {
			t.Error("HandleKey should not handle keys when hidden")
		}
		if chosen {
			t.Error("HandleKey should not choose when hidden")
		}
	})

	t.Run("Render", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()
		c.selection = 0

		output := c.Render()

		// Check that all options are present
		if !strings.Contains(output, "Save & Exit") {
			t.Error("Render should contain 'Save & Exit' option")
		}
		if !strings.Contains(output, "Exit without saving") {
			t.Error("Render should contain 'Exit without saving' option")
		}
		if !strings.Contains(output, "Cancel") {
			t.Error("Render should contain 'Cancel' option")
		}

		// Check that title is present
		if !strings.Contains(output, "Exit editor?") {
			t.Error("Render should contain title 'Exit editor?'")
		}

		// Check that help text is present
		if !strings.Contains(output, "[j/k to select") {
			t.Error("Render should contain help text")
		}
	})

	t.Run("Render with different selections", func(t *testing.T) {
		c := NewExitConfirmation()
		c.Show()

		// Test each selection
		selections := []string{
			"Save & Exit",
			"Exit without saving",
			"Cancel",
		}

		for i, expectedOption := range selections {
			c.selection = i
			output := c.Render()

			// The selected option should have ">" prefix (after styling)
			// We can't easily test for styled content, so just check the option is there
			if !strings.Contains(output, expectedOption) {
				t.Errorf("Selection %d should contain %q", i, expectedOption)
			}
		}
	})
}
