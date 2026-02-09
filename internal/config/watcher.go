package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	// debounceDelay is the time to wait after the last file event before
	// triggering a reload. Editors like vim do rename+write which produces
	// multiple events in quick succession.
	debounceDelay = 300 * time.Millisecond
)

// Watcher watches the config file for changes and invokes a callback
// with the newly loaded configuration. It debounces rapid writes (e.g.
// from vim's rename-and-write pattern) and gracefully handles invalid
// JSON by keeping the previous configuration.
type Watcher struct {
	path     string
	onChange func(*Config)
	watcher  *fsnotify.Watcher
	done     chan struct{}
	once     sync.Once
	logger   *slog.Logger
}

// NewWatcher creates a Watcher that monitors path for changes.
// onChange is called with a freshly loaded Config when the file changes.
// The callback is never called with nil.
func NewWatcher(path string, onChange func(*Config), logger *slog.Logger) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &Watcher{
		path:     path,
		onChange: onChange,
		watcher:  fw,
		done:     make(chan struct{}),
		logger:   logger,
	}, nil
}

// Start begins watching the config file in a background goroutine.
// It watches the parent directory (not the file directly) so that
// editor rename-and-write patterns are detected.
func (w *Watcher) Start() error {
	// Watch the parent directory so we catch rename-write patterns
	dir := filepath.Dir(w.path)
	if err := w.watcher.Add(dir); err != nil {
		return err
	}

	go w.loop()
	return nil
}

// Stop terminates the watcher goroutine and releases resources.
// Safe to call multiple times.
func (w *Watcher) Stop() {
	w.once.Do(func() {
		close(w.done)
		_ = w.watcher.Close()
	})
}

// loop is the main event loop that debounces file system events
// and triggers config reloads.
func (w *Watcher) loop() {
	var timer *time.Timer
	base := filepath.Base(w.path)

	for {
		select {
		case <-w.done:
			if timer != nil {
				timer.Stop()
			}
			return

		case ev, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only react to events on our config file
			if filepath.Base(ev.Name) != base {
				continue
			}

			// We care about writes, creates (rename pattern), and chmod
			if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) == 0 {
				continue
			}

			// Debounce: reset timer on each event
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(debounceDelay, func() {
				w.reload()
			})

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Warn("config watcher error", "error", err)
		}
	}
}

// reload loads the config file and calls onChange if successful.
func (w *Watcher) reload() {
	data, err := os.ReadFile(w.path)
	if err != nil {
		w.logger.Warn("config watcher: cannot read file", "error", err)
		return
	}

	// Quick JSON validity check before full Load (avoids migration side-effects
	// on partial writes)
	if !json.Valid(data) {
		w.logger.Warn("config watcher: invalid JSON, keeping current config")
		return
	}

	cfg, err := LoadFrom(w.path)
	if err != nil {
		w.logger.Warn("config watcher: load failed", "error", err)
		return
	}

	w.logger.Info("config reloaded from disk")
	w.onChange(cfg)
}
