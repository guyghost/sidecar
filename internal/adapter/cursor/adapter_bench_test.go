package cursor

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/guyghost/sidecar/internal/adapter/testutil"
)

// BenchmarkMessages_FullParse measures full message parsing performance.
func BenchmarkMessages_FullParse(b *testing.B) {
	b.Run("Small_10Msgs", func(b *testing.B) {
		benchmarkMessages(b, 10, 512)
	})

	b.Run("Medium_50Msgs", func(b *testing.B) {
		benchmarkMessages(b, 50, 1024)
	})
}

func benchmarkMessages(b *testing.B, messageCount int, avgSize int) {
	tmpDir := b.TempDir()

	// Create project directory and compute workspace hash
	projectRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		b.Fatalf("mkdir project: %v", err)
	}
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		b.Fatalf("abs path: %v", err)
	}
	hash := md5.Sum([]byte(absPath))
	workspaceDir := filepath.Join(tmpDir, "chats", hex.EncodeToString(hash[:]))

	// Create session directory
	sessionID := "bench-session-001"
	sessionDir := filepath.Join(workspaceDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		b.Fatalf("mkdir session: %v", err)
	}

	// Generate test database
	dbPath, err := testutil.GenerateCursorDB(sessionDir, messageCount, avgSize)
	if err != nil {
		b.Fatalf("generate db: %v", err)
	}

	info, _ := os.Stat(dbPath)
	b.Logf("Generated DB size: %d bytes", info.Size())

	a := New()
	a.chatsDir = filepath.Join(tmpDir, "chats")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear session cache to force full parse
		a.sessionCache = make(map[string]sessionCacheEntry)

		_, err := a.Messages(sessionID)
		if err != nil {
			b.Fatalf("Messages failed: %v", err)
		}
	}
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

func benchmarkSessions(b *testing.B, numSessions int) {
	tmpDir := b.TempDir()

	// Create project directory and compute workspace hash
	projectRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		b.Fatalf("mkdir project: %v", err)
	}
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		b.Fatalf("abs path: %v", err)
	}
	hash := md5.Sum([]byte(absPath))
	workspaceDir := filepath.Join(tmpDir, "chats", hex.EncodeToString(hash[:]))

	// Create multiple session directories
	for i := 0; i < numSessions; i++ {
		sessionID := filepath.Join("bench-session", string(rune('A'+i%26)), string(rune('0'+i%10)))
		sessionDir := filepath.Join(workspaceDir, sessionID)
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			b.Fatalf("mkdir session %d: %v", i, err)
		}

		_, err := testutil.GenerateCursorDB(sessionDir, 10, 512)
		if err != nil {
			b.Fatalf("generate db %d: %v", i, err)
		}
	}

	a := New()
	a.chatsDir = filepath.Join(tmpDir, "chats")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear session cache between runs
		a.sessionCache = make(map[string]sessionCacheEntry)

		_, err := a.Sessions(projectRoot)
		if err != nil {
			b.Fatalf("Sessions failed: %v", err)
		}
	}
}

// BenchmarkMessages_Allocs specifically tracks allocations for optimization.
func BenchmarkMessages_Allocs(b *testing.B) {
	tmpDir := b.TempDir()

	// Create project directory and compute workspace hash
	projectRoot := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(projectRoot, 0755); err != nil {
		b.Fatalf("mkdir project: %v", err)
	}
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		b.Fatalf("abs path: %v", err)
	}
	hash := md5.Sum([]byte(absPath))
	workspaceDir := filepath.Join(tmpDir, "chats", hex.EncodeToString(hash[:]))

	// Create session directory
	sessionID := "bench-session-001"
	sessionDir := filepath.Join(workspaceDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		b.Fatalf("mkdir session: %v", err)
	}

	// Generate small database for allocation focus
	_, err = testutil.GenerateCursorDB(sessionDir, 20, 256)
	if err != nil {
		b.Fatalf("generate db: %v", err)
	}

	a := New()
	a.chatsDir = filepath.Join(tmpDir, "chats")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		a.sessionCache = make(map[string]sessionCacheEntry)
		_, _ = a.Messages(sessionID)
	}
}
