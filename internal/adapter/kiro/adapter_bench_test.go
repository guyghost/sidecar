package kiro

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/guyghost/sidecar/internal/adapter/testutil"
)

// BenchmarkMessages_FullParse measures full message parsing performance.
func BenchmarkMessages_FullParse(b *testing.B) {
	b.Run("Small_10Msgs", func(b *testing.B) {
		benchmarkMessages(b, 1, 10, 512)
	})

	b.Run("Medium_50Msgs", func(b *testing.B) {
		benchmarkMessages(b, 1, 50, 1024)
	})

	b.Run("Large_200Msgs", func(b *testing.B) {
		if testing.Short() {
			b.Skip("skipping large benchmark in short mode")
		}
		benchmarkMessages(b, 1, 200, 1024)
	})
}

func benchmarkMessages(b *testing.B, convCount int, msgsPerConv int, avgSize int) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "kiro.sqlite3")
	projectPath := filepath.Join(tmpDir, "project")

	// Generate test database
	if err := testutil.GenerateKiroDB(dbPath, projectPath, convCount, msgsPerConv, avgSize); err != nil {
		b.Fatalf("generate db: %v", err)
	}

	info, _ := os.Stat(dbPath)
	b.Logf("Generated DB size: %d bytes", info.Size())

	a := New()
	a.dbPath = dbPath

	// Get a valid conversation ID from Sessions
	sessions, err := a.Sessions(projectPath)
	if err != nil || len(sessions) == 0 {
		b.Fatalf("no sessions found: %v", err)
	}
	sessionID := sessions[0].ID
	b.Logf("Benchmarking with session ID: %s", sessionID)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Keep DB connection for Messages benchmarks (no reconnect)

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
	b.Run("n=5", func(b *testing.B) {
		benchmarkSessions(b, 5)
	})

	b.Run("n=20", func(b *testing.B) {
		benchmarkSessions(b, 20)
	})
}

func benchmarkSessions(b *testing.B, convCount int) {
	tmpDir := b.TempDir()
	dbPath := filepath.Join(tmpDir, "kiro.sqlite3")
	projectPath := filepath.Join(tmpDir, "project")

	// Generate test database
	if err := testutil.GenerateKiroDB(dbPath, projectPath, convCount, 10, 512); err != nil {
		b.Fatalf("generate db: %v", err)
	}

	a := New()
	a.dbPath = dbPath

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Force DB reconnect for Sessions benchmarks
		if a.db != nil {
			_ = a.Close()
			a.db = nil
		}

		_, err := a.Sessions(projectPath)
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
	dbPath := filepath.Join(tmpDir, "kiro.sqlite3")
	projectPath := filepath.Join(tmpDir, "project")

	// Generate small database for allocation focus
	if err := testutil.GenerateKiroDB(dbPath, projectPath, 1, 20, 256); err != nil {
		b.Fatalf("generate db: %v", err)
	}

	a := New()
	a.dbPath = dbPath

	// Get a valid conversation ID
	sessions, err := a.Sessions(projectPath)
	if err != nil || len(sessions) == 0 {
		b.Fatalf("no sessions found: %v", err)
	}
	sessionID := sessions[0].ID

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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
