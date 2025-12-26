package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// State holds persistent user preferences.
type State struct {
	GitDiffMode string `json:"gitDiffMode"` // "unified" or "side-by-side"
}

var (
	current *State
	mu      sync.RWMutex
	path    string
)

// Init loads state from the default location.
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path = filepath.Join(home, ".config", "sidecar", "state.json")
	return Load()
}

// Load reads state from disk.
func Load() error {
	mu.Lock()
	defer mu.Unlock()

	current = &State{
		GitDiffMode: "unified", // default
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // no state file yet, use defaults
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, current)
}

// Save writes state to disk.
func Save() error {
	mu.RLock()
	defer mu.RUnlock()

	if current == nil {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetGitDiffMode returns the saved diff mode.
func GetGitDiffMode() string {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		return "unified"
	}
	return current.GitDiffMode
}

// SetGitDiffMode saves the diff mode preference.
func SetGitDiffMode(mode string) error {
	mu.Lock()
	if current == nil {
		current = &State{}
	}
	current.GitDiffMode = mode
	mu.Unlock()
	return Save()
}
