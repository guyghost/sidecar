package perf

import (
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	snap := Collect()

	if snap.Goroutines <= 0 {
		t.Errorf("expected positive goroutine count, got %d", snap.Goroutines)
	}
	if snap.HeapAlloc == 0 {
		t.Error("expected non-zero HeapAlloc")
	}
	if snap.HeapSys == 0 {
		t.Error("expected non-zero HeapSys")
	}
	if snap.Uptime < 0 {
		t.Errorf("expected non-negative uptime, got %s", snap.Uptime)
	}
	// FDs can be -1 on unsupported platforms — that's OK
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{10485760, "10.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.input)
		if got != tt.want {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{0, "0s"},
		{5 * time.Second, "5s"},
		{90 * time.Second, "1m 30s"},
		{5 * time.Minute, "5m 0s"},
		{65 * time.Minute, "1h 5m"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
	}

	for _, tt := range tests {
		got := FormatUptime(tt.input)
		if got != tt.want {
			t.Errorf("FormatUptime(%s) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
