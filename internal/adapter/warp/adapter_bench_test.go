package warp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guyghost/sidecar/internal/adapter/testutil"
)

// BenchmarkMessages_FullParse measures full message parsing performance.
func BenchmarkMessages_FullParse(b *testing.B) {
	b.Run("Small_5Conv_10Msgs", func(b *testing.B) {
		benchmarkMessages(b, 5, 10, 512)
	})

	b.Run("Medium_20Conv_50Msgs", func(b *testing.B) {
		if testing.Short() {
			b.Skip("skipping medium benchmark in short mode")
		}
		benchmarkMessages(b, 20, 50, 1024)
	})
}

func benchmarkMessages(b *testing.B, convCount int, msgsPerConv int, avgSize int) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "warp.sqlite")
	projectDir := filepath.Join(tmpDir, "project")

	// Generate test database
	if err := testutil.GenerateWarpDB(dbPath, convCount, msgsPerConv, projectDir); err != nil {
		b.Fatalf("generate db: %v", err)
	}

	info, _ := os.Stat(dbPath)
	b.Logf("Generated DB size: %d bytes", info.Size())

	a := New()
	a.dbPath = dbPath

	// Get a valid conversation ID from Sessions
	sessions, err := a.Sessions(projectDir)
	if err != nil || len(sessions) == 0 {
		b.Fatalf("no sessions found: %v", err)
	}
	sessionID := sessions[0].ID
	b.Logf("Benchmarking with session ID: %s", sessionID)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear session index, but keep DB connection
		a.sessionIndex = make(map[string]struct{})

		_, err := a.Messages(sessionID)
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}

	// Cleanup: close DB connection
	b.Cleanup(func() {
		if a.db != nil {
			_ = a.Close()
		}
	})
}

// BenchmarkSessions measures session discovery performance.
func BenchmarkSessions(b *testing.B) {
	b.Run("n=10", func(b *testing.B) {
		benchmarkSessions(b, 10)
	})

	b.Run("n=50", func(b *testing.B) {
		if testing.Short() {
			b.Skip("skipping 50 sessions benchmark in short mode")
		}
		benchmarkSessions(b, 50)
	})
}

func benchmarkSessions(b *testing.B, convCount int) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "warp.sqlite")
	projectDir := filepath.Join(tmpDir, "project")

	// Generate test database
	if err := testutil.GenerateWarpDB(dbPath, convCount, 10, projectDir); err != nil {
		b.Fatalf("generate db: %v", err)
	}

	a := New()
	a.dbPath = dbPath

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear session index and force DB reconnect
		a.sessionIndex = make(map[string]struct{})

		// Close DB to force reconnect for this benchmark
		if a.db != nil {
			_ = a.Close()
			a.db = nil
		}

		_, err := a.Sessions(projectDir)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}

	// Cleanup: close DB connection
	b.Cleanup(func() {
		if a.db != nil {
			_ = a.Close()
		}
		// Clean up temp DB file
		_ = os.Remove(dbPath)
	})
}

// BenchmarkMessages_Allocs specifically tracks allocations for optimization.
func BenchmarkMessages_Allocs(b *testing.B) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "warp.sqlite")
	projectDir := filepath.Join(tmpDir, "project")

	// Generate small database for allocation focus
	if err := testutil.GenerateWarpDB(dbPath, 5, 5, projectDir); err != nil {
		b.Fatalf("generate db: %v", err)
	}

	a := New()
	a.dbPath = dbPath

	// Get a valid conversation ID
	sessions, err := a.Sessions(projectDir)
	if err != nil || len(sessions) == 0 {
		b.Fatalf("no sessions found: %v", err)
	}
	sessionID := sessions[0].ID

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.sessionIndex = make(map[string]struct{})
		_, _ = a.Messages(sessionID)
	}

	// Cleanup: close DB connection
	b.Cleanup(func() {
		if a.db != nil {
			_ = a.Close()
		}
		_ = os.Remove(dbPath)
	})
}
