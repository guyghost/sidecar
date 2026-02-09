package adapter

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestPartialResult_Error(t *testing.T) {
	pr := &PartialResult{
		Err:         io.ErrUnexpectedEOF,
		ParsedCount: 42,
		Reason:      "truncated JSON",
	}
	msg := pr.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
	// Should contain all three pieces of info.
	for _, want := range []string{"truncated JSON", "42", "unexpected EOF"} {
		if !contains(msg, want) {
			t.Errorf("error message %q missing %q", msg, want)
		}
	}
}

func TestPartialResult_Unwrap(t *testing.T) {
	inner := io.ErrUnexpectedEOF
	pr := &PartialResult{Err: inner, ParsedCount: 1, Reason: "test"}
	if !errors.Is(pr, inner) {
		t.Error("expected errors.Is to match inner error")
	}
}

func TestIsPartial(t *testing.T) {
	t.Run("direct PartialResult", func(t *testing.T) {
		pr := &PartialResult{Err: io.EOF, ParsedCount: 5, Reason: "test"}
		got, ok := IsPartial(pr)
		if !ok {
			t.Fatal("expected IsPartial to return true")
		}
		if got.ParsedCount != 5 {
			t.Errorf("expected ParsedCount=5, got %d", got.ParsedCount)
		}
	})

	t.Run("wrapped PartialResult", func(t *testing.T) {
		pr := &PartialResult{Err: io.EOF, ParsedCount: 3, Reason: "wrapped"}
		wrapped := fmt.Errorf("outer: %w", pr)
		got, ok := IsPartial(wrapped)
		if !ok {
			t.Fatal("expected IsPartial to match wrapped error")
		}
		if got.ParsedCount != 3 {
			t.Errorf("expected ParsedCount=3, got %d", got.ParsedCount)
		}
	})

	t.Run("non-partial error", func(t *testing.T) {
		_, ok := IsPartial(io.EOF)
		if ok {
			t.Error("expected IsPartial to return false for non-partial error")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		_, ok := IsPartial(nil)
		if ok {
			t.Error("expected IsPartial to return false for nil")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
