package gitstatus

import (
	"testing"
)

func TestExternalTool_NewExternalTool(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)
	if tool == nil {
		t.Fatal("NewExternalTool returned nil")
	}
	if tool.Mode() != ToolModeAuto {
		t.Errorf("Mode() = %v, want %v", tool.Mode(), ToolModeAuto)
	}
}

func TestExternalTool_SetMode(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)
	tool.SetMode(ToolModeBuiltin)
	if tool.Mode() != ToolModeBuiltin {
		t.Errorf("Mode() = %v, want %v", tool.Mode(), ToolModeBuiltin)
	}
}

func TestExternalTool_ShouldUseDelta_Builtin(t *testing.T) {
	tool := NewExternalTool(ToolModeBuiltin)
	if tool.ShouldUseDelta() {
		t.Error("ShouldUseDelta() = true for ToolModeBuiltin, want false")
	}
}

func TestExternalTool_ShouldUseDelta_DeltaMode_NoDelta(t *testing.T) {
	// Force delta to not be found by using a mode where it matters
	tool := NewExternalTool(ToolModeDelta)
	// If delta is not installed, this should return false
	// The actual behavior depends on whether delta is installed
	// This test just ensures no panic
	_ = tool.ShouldUseDelta()
}

func TestExternalTool_ShouldShowTip_OnlyOnce(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)

	// Force delta to not be found
	tool.deltaPath = "" // Ensure delta is not "found"

	// First call should return true (if delta not installed)
	if !tool.HasDelta() {
		first := tool.ShouldShowTip()
		second := tool.ShouldShowTip()

		if !first {
			t.Error("first ShouldShowTip() = false, want true")
		}
		if second {
			t.Error("second ShouldShowTip() = true, want false")
		}
	}
}

func TestExternalTool_ShouldShowTip_WithDelta(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)

	// Force delta to be "found"
	tool.deltaPath = "/usr/local/bin/delta"

	// Should not show tip when delta is installed
	if tool.ShouldShowTip() {
		t.Error("ShouldShowTip() = true when delta installed, want false")
	}
}

func TestExternalTool_GetTipMessage(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)
	msg := tool.GetTipMessage()

	if msg == "" {
		t.Error("GetTipMessage() returned empty string")
	}
	if len(msg) < 10 {
		t.Error("GetTipMessage() returned unexpectedly short message")
	}
}

func TestExternalTool_RenderWithDelta_NoDelta(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)
	tool.deltaPath = "" // No delta installed

	input := "+added line\n-removed line"
	output, err := tool.RenderWithDelta(input, false, 80)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Without delta, should return input unchanged
	if output != input {
		t.Errorf("output = %q, want %q", output, input)
	}
}

func TestExternalTool_RenderWithDelta_SideBySide_NoDelta(t *testing.T) {
	tool := NewExternalTool(ToolModeAuto)
	tool.deltaPath = "" // No delta installed

	input := "+added line\n-removed line"
	output, err := tool.RenderWithDelta(input, true, 120)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Without delta, should return input unchanged
	if output != input {
		t.Errorf("output = %q, want %q", output, input)
	}
}

func TestExternalToolMode_Constants(t *testing.T) {
	// Verify the constants exist and are distinct
	modes := []ExternalToolMode{ToolModeAuto, ToolModeDelta, ToolModeBuiltin}
	seen := make(map[ExternalToolMode]bool)

	for _, m := range modes {
		if seen[m] {
			t.Errorf("duplicate mode value: %v", m)
		}
		seen[m] = true
	}
}
