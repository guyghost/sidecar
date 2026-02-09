package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWatcher_BasicReload(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	// Write initial config
	initial := `{"ui":{"showClock":true}}`
	if err := os.WriteFile(cfgPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var received *Config

	w, err := NewWatcher(cfgPath, func(cfg *Config) {
		mu.Lock()
		received = cfg
		mu.Unlock()
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// Point loader at our temp config
	SetTestConfigPath(cfgPath)
	defer ResetTestConfigPath()

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Modify config
	updated := `{"ui":{"showClock":false}}`
	if err := os.WriteFile(cfgPath, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for debounce + processing
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := received
		mu.Unlock()
		if got != nil {
			if got.UI.ShowClock {
				t.Error("expected ShowClock=false after reload")
			}
			return // success
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for config reload")
}

func TestWatcher_DebounceRapidWrites(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	initial := `{"ui":{"showClock":true}}`
	if err := os.WriteFile(cfgPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	var callCount atomic.Int32

	w, err := NewWatcher(cfgPath, func(cfg *Config) {
		callCount.Add(1)
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	SetTestConfigPath(cfgPath)
	defer ResetTestConfigPath()

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Rapid writes — should debounce to a single callback
	for i := 0; i < 10; i++ {
		cfg := map[string]any{"ui": map[string]any{"showClock": i%2 == 0}}
		data, _ := json.Marshal(cfg)
		if err := os.WriteFile(cfgPath, data, 0644); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond) // faster than debounce (300ms)
	}

	// Wait for debounce to settle
	time.Sleep(600 * time.Millisecond)

	count := callCount.Load()
	if count == 0 {
		t.Fatal("expected at least one callback")
	}
	if count > 3 {
		// With 300ms debounce and 10 writes at 20ms intervals,
		// we should get at most 1-2 callbacks, be lenient for CI
		t.Errorf("expected debouncing to limit callbacks, got %d", count)
	}
}

func TestWatcher_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	initial := `{"ui":{"showClock":true}}`
	if err := os.WriteFile(cfgPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	var callCount atomic.Int32

	w, err := NewWatcher(cfgPath, func(cfg *Config) {
		callCount.Add(1)
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	SetTestConfigPath(cfgPath)
	defer ResetTestConfigPath()

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Write invalid JSON
	if err := os.WriteFile(cfgPath, []byte(`{invalid json`), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for debounce + processing
	time.Sleep(600 * time.Millisecond)

	if callCount.Load() != 0 {
		t.Error("callback should not fire for invalid JSON")
	}
}

func TestWatcher_VimStyleRename(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	initial := `{"ui":{"showClock":true}}`
	if err := os.WriteFile(cfgPath, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	var mu sync.Mutex
	var received *Config

	w, err := NewWatcher(cfgPath, func(cfg *Config) {
		mu.Lock()
		received = cfg
		mu.Unlock()
	}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	SetTestConfigPath(cfgPath)
	defer ResetTestConfigPath()

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Simulate vim: write to temp, rename over original
	tmpPath := cfgPath + ".tmp"
	updated := `{"ui":{"showClock":false}}`
	if err := os.WriteFile(tmpPath, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tmpPath, cfgPath); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		got := received
		mu.Unlock()
		if got != nil {
			if got.UI.ShowClock {
				t.Error("expected ShowClock=false after vim-style save")
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for config reload after vim-style save")
}

func TestWatcher_StopCleanup(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	if err := os.WriteFile(cfgPath, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	w, err := NewWatcher(cfgPath, func(cfg *Config) {}, slog.Default())
	if err != nil {
		t.Fatal(err)
	}

	if err := w.Start(); err != nil {
		t.Fatal(err)
	}

	// Stop should be safe to call multiple times
	w.Stop()
	w.Stop() // second call should not panic
}
