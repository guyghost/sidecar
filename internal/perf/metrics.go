// Package perf provides lightweight performance counters for diagnostics.
// All functions are safe for concurrent use and designed for periodic sampling.
package perf

import (
	"runtime"
	"time"

	"github.com/guyghost/sidecar/internal/fdmonitor"
)

// startTime records when the process started for uptime calculation.
var startTime = time.Now()

// Snapshot holds a point-in-time performance sample.
type Snapshot struct {
	Goroutines int
	HeapAlloc  uint64 // bytes currently allocated on the heap
	HeapSys    uint64 // bytes obtained from the OS for heap
	NumGC      uint32 // completed GC cycles
	FDs        int    // open file descriptors (-1 if unavailable)
	Uptime     time.Duration
}

// Collect gathers a performance snapshot from the Go runtime and OS.
func Collect() Snapshot {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return Snapshot{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  mem.HeapAlloc,
		HeapSys:    mem.HeapSys,
		NumGC:      mem.NumGC,
		FDs:        fdmonitor.Count(),
		Uptime:     time.Since(startTime).Truncate(time.Second),
	}
}

// FormatBytes returns a human-readable byte size (e.g. "12.3 MB").
func FormatBytes(b uint64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return formatFloat(float64(b)/float64(gb)) + " GB"
	case b >= mb:
		return formatFloat(float64(b)/float64(mb)) + " MB"
	case b >= kb:
		return formatFloat(float64(b)/float64(kb)) + " KB"
	default:
		return formatUint(b) + " B"
	}
}

// FormatUptime returns uptime as "Xh Ym Zs" or shorter forms.
func FormatUptime(d time.Duration) string {
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	switch {
	case h > 0:
		return formatInt(h) + "h " + formatInt(m) + "m"
	case m > 0:
		return formatInt(m) + "m " + formatInt(s) + "s"
	default:
		return formatInt(s) + "s"
	}
}

// formatFloat returns a float with 1 decimal place without importing fmt.
func formatFloat(f float64) string {
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	if frac < 0 {
		frac = -frac
	}
	return formatInt(whole) + "." + formatInt(frac)
}

// formatInt returns an integer as a string without importing fmt.
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

// formatUint returns a uint64 as a string.
func formatUint(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
