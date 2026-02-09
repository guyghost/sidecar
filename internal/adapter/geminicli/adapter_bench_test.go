package geminicli

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
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// Small: 10 messages (~100KB)
	sessionFile := filepath.Join(chatsDir, "session-001.json")
	if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 10, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.tmpDir = tmpDir
	a.sessionIndex["session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache to force full parse
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages("session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_FullParse_Medium(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// Medium: 100 messages (~1MB)
	sessionFile := filepath.Join(chatsDir, "session-001.json")
	if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 100, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	info, _ := os.Stat(sessionFile)
	b.Logf("Generated file size: %d bytes", info.Size())

	a := New()
	a.tmpDir = tmpDir
	a.sessionIndex["session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages("session-001")
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
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// Large: 1000 messages (~10MB)
	sessionFile := filepath.Join(chatsDir, "session-001.json")
	if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 1000, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	info, _ := os.Stat(sessionFile)
	b.Logf("Generated file size: %d bytes", info.Size())

	a := New()
	a.tmpDir = tmpDir
	a.sessionIndex["session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages("session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_CacheHit(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// ~1MB file
	sessionFile := filepath.Join(chatsDir, "session-001.json")
	if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 100, 1024); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.tmpDir = tmpDir
	a.sessionIndex["session-001"] = sessionFile

	// Warm the cache
	_, err := a.Messages("session-001")
	if err != nil {
		b.Fatalf("initial Messages failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := a.Messages("session-001")
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkSessions_10(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// Create 10 session files
	for i := 0; i < 10; i++ {
		sessionFile := filepath.Join(chatsDir, fmt.Sprintf("session-%d.json", i))
		if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 10, 512); err != nil {
			b.Fatalf("failed to generate test file %d: %v", i, err)
		}
	}

	a := New()
	a.tmpDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.sessionIndex = make(map[string]string)
		a.metaCache = make(map[string]sessionMetaCacheEntry)
		_, err := a.Sessions(projectDir)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

func BenchmarkSessions_50(b *testing.B) {
	tmpDir := b.TempDir()
	projectDir := "/bench/project"
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// Create 50 session files
	for i := 0; i < 50; i++ {
		sessionFile := filepath.Join(chatsDir, fmt.Sprintf("session-%d.json", i))
		if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 10, 512); err != nil {
			b.Fatalf("failed to generate test file %d: %v", i, err)
		}
	}

	a := New()
	a.tmpDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.sessionIndex = make(map[string]string)
		a.metaCache = make(map[string]sessionMetaCacheEntry)
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
	hash := projectHash(projectDir)
	chatsDir := filepath.Join(tmpDir, hash, "chats")
	_ = os.MkdirAll(chatsDir, 0755)

	// 100 messages to focus on allocation patterns
	sessionFile := filepath.Join(chatsDir, "session-001.json")
	if err := testutil.GenerateGeminiCLISessionFile(sessionFile, 100, 256); err != nil {
		b.Fatalf("failed to generate test file: %v", err)
	}

	a := New()
	a.tmpDir = tmpDir
	a.sessionIndex["session-001"] = sessionFile

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, _ = a.Messages("session-001")
	}
}
