package styles

import (
	"testing"
)

func TestBackwardCompatibility(t *testing.T) {
	// Test that all public package-level variables are initialized
	tests := []struct {
		name  string
		value interface{}
	}{
		{"Primary", Primary},
		{"Secondary", Secondary},
		{"Success", Success},
		{"PanelActive", PanelActive},
		{"Title", Title},
		{"Muted", Muted},
		{"StatusBlocked", StatusBlocked},
	}

	for _, tt := range tests {
		if tt.value == nil {
			t.Errorf("%s is not initialized", tt.name)
		}
	}

	// Test that GetStyles() works
	s := GetStyles()
	if s == nil {
		t.Error("GetStyles() returned nil")
	}

	// Test that Current atomic pointer is set
	if Current.Load() == nil {
		t.Error("Current atomic pointer is nil")
	}
}

func TestApplyTheme(t *testing.T) {
	// Test that ApplyTheme creates new styles
	initialPrimary := Primary

	ApplyTheme("dracula")
	if Primary == initialPrimary {
		t.Error("Primary did not change after ApplyTheme")
	}

	// Test that Current is updated
	s := Current.Load()
	if s == nil {
		t.Error("Current is nil after ApplyTheme")
	}
	if s.Primary != Primary {
		t.Error("Current.Primary != Primary after ApplyTheme")
	}

	// Reset to default
	ApplyTheme("default")
}
