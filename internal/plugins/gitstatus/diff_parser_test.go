package gitstatus

import (
	"testing"

	"github.com/marcus/sidecar/internal/git"
)

// Tests for diff parsing live in internal/git/diff_test.go.
// This file verifies that re-exports work correctly.

func TestDiffParserReexports(t *testing.T) {
	diff := `--- a/file.txt
+++ b/file.txt
@@ -1,3 +1,3 @@
 context
-old
+new
 more context`

	// Use the re-exported function
	parsed, err := ParseUnifiedDiff(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify types are compatible
	var _ *git.ParsedDiff = parsed
	if len(parsed.Hunks) != 1 {
		t.Fatalf("len(Hunks) = %d, want 1", len(parsed.Hunks))
	}

	// Verify constants match
	if LineAdd != git.LineAdd {
		t.Error("LineAdd constant mismatch")
	}
	if LineRemove != git.LineRemove {
		t.Error("LineRemove constant mismatch")
	}
	if LineContext != git.LineContext {
		t.Error("LineContext constant mismatch")
	}
}
