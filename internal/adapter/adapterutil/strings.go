package adapterutil

import (
	"strings"
)

// ShortID returns the first 8 characters of an ID, or the full ID if shorter.
// This provides a consistent, short identifier for display purposes.
func ShortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

// TruncateTitle truncates text to maxLen, adding "..." if truncated.
// It also replaces newlines with spaces for display.
func TruncateTitle(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.TrimSpace(s)

	if len(s) <= maxLen {
		return s
	}

	// If maxLen is too small to include "..." just truncate without ellipsis
	if maxLen < 4 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}
