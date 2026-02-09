package amp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/guyghost/sidecar/internal/adapter/cache"
	"github.com/guyghost/sidecar/internal/adapter/testutil"
)

// Benchmark targets (td-336ee0):
// - Full parse (1MB): <50ms
// - Full parse (10MB): <500ms
// - Cache hit: <1ms

func BenchmarkMessages_FullParse_Small(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// Small: 10 messages (~100KB)
	threadFile := filepath.Join(tmpDir, "T-bench001.json")
	if err := testutil.GenerateAmpThreadFile(threadFile, 10, 1024, projectURI); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.threadsDir = tmpDir
	a.sessionIndex["T-bench001"] = threadFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache to force full parse
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages("T-bench001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_FullParse_Medium(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// Medium: 100 messages (~1MB)
	threadFile := filepath.Join(tmpDir, "T-bench001.json")
	if err := testutil.GenerateAmpThreadFile(threadFile, 100, 1024, projectURI); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	info, _ := os.Stat(threadFile)
	b.Logf("Generated file size: %d bytes", info.Size())

	a := New()
	a.threadsDir = tmpDir
	a.sessionIndex["T-bench001"] = threadFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages("T-bench001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_FullParse_Large(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping large benchmark in short mode")
	}

	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// Large: 1000 messages (~10MB)
	threadFile := filepath.Join(tmpDir, "T-bench001.json")
	if err := testutil.GenerateAmpThreadFile(threadFile, 1000, 1024, projectURI); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	info, _ := os.Stat(threadFile)
	b.Logf("Generated file size: %d bytes", info.Size())

	a := New()
	a.threadsDir = tmpDir
	a.sessionIndex["T-bench001"] = threadFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages("T-bench001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_CacheHit(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// ~1MB file
	threadFile := filepath.Join(tmpDir, "T-bench001.json")
	if err := testutil.GenerateAmpThreadFile(threadFile, 100, 1024, projectURI); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.threadsDir = tmpDir
	a.sessionIndex["T-bench001"] = threadFile

	// Warm the cache
	_, err := a.Messages("T-bench001")
	if err != nil {
		b.Fatalf("initial Messages failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := a.Messages("T-bench001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkSessions_10(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// Create 10 thread files
	for i := 0; i < 10; i++ {
		threadFile := filepath.Join(tmpDir, fmt.Sprintf("T-bench%03d.json", i))
		if err := testutil.GenerateAmpThreadFile(threadFile, 10, 512, projectURI); err != nil {
			b.Fatalf("failed to generate test file %d: %v", i, err)
		}
	}

	a := New()
	a.threadsDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.sessionIndex = make(map[string]string)
		a.metaCache = make(map[string]metaCacheEntry)
		a.projectMatchCache = cache.New[projectMatchCacheEntry](100)
		_, err := a.Sessions(projectDir)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

func BenchmarkSessions_50(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// Create 50 thread files
	for i := 0; i < 50; i++ {
		threadFile := filepath.Join(tmpDir, fmt.Sprintf("T-bench%03d.json", i))
		if err := testutil.GenerateAmpThreadFile(threadFile, 10, 512, projectURI); err != nil {
			b.Fatalf("failed to generate test file %d: %v", i, err)
		}
	}

	a := New()
	a.threadsDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.sessionIndex = make(map[string]string)
		a.metaCache = make(map[string]metaCacheEntry)
		a.projectMatchCache = cache.New[projectMatchCacheEntry](100)
		_, err := a.Sessions(projectDir)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

// BenchmarkMessages_Allocs specifically tracks allocations for optimization.
func BenchmarkMessages_Allocs(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	projectURI := "file://" + projectDir

	// 100 messages to focus on allocation patterns
	threadFile := filepath.Join(tmpDir, "T-bench001.json")
	if err := testutil.GenerateAmpThreadFile(threadFile, 100, 256, projectURI); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.threadsDir = tmpDir
	a.sessionIndex["T-bench001"] = threadFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, _ = a.Messages("T-bench001")
	}
}
