package opencode

import (
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
	projectPath := filepath.Join(tmpDir, "project")

	// Create the project directory so the worktree path is valid
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	// Small: 5 sessions, 10 messages each
	if err := testutil.GenerateOpenCodeStorage(tmpDir, projectPath, 5, 10, 1024); err != nil {
		b.Fatalf("failed to generate test storage: %v", err)
	}

	a := New()
	a.storageDir = tmpDir

	// Call Sessions to populate index and get a valid sessionID
	sessions, err := a.Sessions(projectPath)
	if err != nil {
		b.Fatalf("Sessions failed: %v", err)
	}
	if len(sessions) == 0 {
		b.Fatal("no sessions generated")
	}
	sessionID := sessions[0].ID

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear cache to force full parse
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages(sessionID)
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_FullParse_Medium(b *testing.B) {
	tmpDir := b.TempDir()
	projectPath := filepath.Join(tmpDir, "project")

	// Create the project directory so the worktree path is valid
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	// Medium: 5 sessions, 100 messages each (~1MB per session)
	if err := testutil.GenerateOpenCodeStorage(tmpDir, projectPath, 5, 100, 1024); err != nil {
		b.Fatalf("failed to generate test storage: %v", err)
	}

	a := New()
	a.storageDir = tmpDir

	// Call Sessions to populate index and get a valid sessionID
	sessions, err := a.Sessions(projectPath)
	if err != nil {
		b.Fatalf("Sessions failed: %v", err)
	}
	if len(sessions) == 0 {
		b.Fatal("no sessions generated")
	}
	sessionID := sessions[0].ID

	// Count message directory files for logging
	messageDir := filepath.Join(tmpDir, "message", sessionID)
	entries, _ := os.ReadDir(messageDir)
	b.Logf("Generated %d messages for session %s", len(entries), sessionID)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, err := a.Messages(sessionID)
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkMessages_CacheHit(b *testing.B) {
	tmpDir := b.TempDir()
	projectPath := filepath.Join(tmpDir, "project")

	// Create the project directory so the worktree path is valid
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	// Generate storage with 5 sessions, 50 messages each
	if err := testutil.GenerateOpenCodeStorage(tmpDir, projectPath, 5, 50, 1024); err != nil {
		b.Fatalf("failed to generate test storage: %v", err)
	}

	a := New()
	a.storageDir = tmpDir

	// Call Sessions to populate index and get a valid sessionID
	sessions, err := a.Sessions(projectPath)
	if err != nil {
		b.Fatalf("Sessions failed: %v", err)
	}
	if len(sessions) == 0 {
		b.Fatal("no sessions generated")
	}
	sessionID := sessions[0].ID

	// Warm the cache
	_, err = a.Messages(sessionID)
	if err != nil {
		b.Fatalf("initial Messages failed: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := a.Messages(sessionID)
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
}

func BenchmarkSessions_5(b *testing.B) {
	tmpDir := b.TempDir()
	projectPath := filepath.Join(tmpDir, "project")

	// Create the project directory so the worktree path is valid
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	// 5 sessions with 10 messages each
	if err := testutil.GenerateOpenCodeStorage(tmpDir, projectPath, 5, 10, 512); err != nil {
		b.Fatalf("failed to generate test storage: %v", err)
	}

	a := New()
	a.storageDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.projectIndex = nil
		a.sessionIndex = nil
		a.projectsLoaded = false
		a.metaCache = make(map[string]sessionMetaCacheEntry)
		a.msgCache = nil
		_, err := a.Sessions(projectPath)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

func BenchmarkSessions_20(b *testing.B) {
	tmpDir := b.TempDir()
	projectPath := filepath.Join(tmpDir, "project")

	// Create the project directory so the worktree path is valid
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	// 20 sessions with 10 messages each
	if err := testutil.GenerateOpenCodeStorage(tmpDir, projectPath, 20, 10, 512); err != nil {
		b.Fatalf("failed to generate test storage: %v", err)
	}

	a := New()
	a.storageDir = tmpDir

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear caches
		a.projectIndex = nil
		a.sessionIndex = nil
		a.projectsLoaded = false
		a.metaCache = make(map[string]sessionMetaCacheEntry)
		a.msgCache = nil
		_, err := a.Sessions(projectPath)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

// BenchmarkMessages_Allocs specifically tracks allocations for optimization.
func BenchmarkMessages_Allocs(b *testing.B) {
	tmpDir := b.TempDir()
	projectPath := filepath.Join(tmpDir, "project")

	// Create the project directory so the worktree path is valid
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		b.Fatalf("failed to create project directory: %v", err)
	}

	// Generate storage with 5 sessions, 100 messages each
	if err := testutil.GenerateOpenCodeStorage(tmpDir, projectPath, 5, 100, 256); err != nil {
		b.Fatalf("failed to generate test storage: %v", err)
	}

	a := New()
	a.storageDir = tmpDir

	// Call Sessions to populate index and get a valid sessionID
	sessions, err := a.Sessions(projectPath)
	if err != nil {
		b.Fatalf("Sessions failed: %v", err)
	}
	if len(sessions) == 0 {
		b.Fatal("no sessions generated")
	}
	sessionID := sessions[0].ID

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.msgCache = cache.New[msgCacheEntry](100)
		_, _ = a.Messages(sessionID)
	}
}
