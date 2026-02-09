package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]interface{}
		want int
	}{
		{
			name: "no version field → v1",
			raw:  map[string]interface{}{"ui": map[string]interface{}{}},
			want: 1,
		},
		{
			name: "version 1",
			raw:  map[string]interface{}{"version": float64(1)},
			want: 1,
		},
		{
			name: "version 2",
			raw:  map[string]interface{}{"version": float64(2)},
			want: 2,
		},
		{
			name: "version 0 clamped to 1",
			raw:  map[string]interface{}{"version": float64(0)},
			want: 1,
		},
		{
			name: "negative version clamped to 1",
			raw:  map[string]interface{}{"version": float64(-5)},
			want: 1,
		},
		{
			name: "future version clamped to CurrentVersion",
			raw:  map[string]interface{}{"version": float64(999)},
			want: CurrentVersion,
		},
		{
			name: "non-numeric version → v1",
			raw:  map[string]interface{}{"version": "two"},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectVersion(tt.raw)
			if got != tt.want {
				t.Errorf("detectVersion() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMigrateV1ToV2(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		wantCommunity string
		wantOverrides bool // whether overrides should still exist
	}{
		{
			name: "moves communityName to community",
			input: map[string]interface{}{
				"ui": map[string]interface{}{
					"theme": map[string]interface{}{
						"name": "default",
						"overrides": map[string]interface{}{
							"communityName": "catppuccin",
						},
					},
				},
			},
			wantCommunity: "catppuccin",
			wantOverrides: false,
		},
		{
			name: "preserves other overrides",
			input: map[string]interface{}{
				"ui": map[string]interface{}{
					"theme": map[string]interface{}{
						"name": "default",
						"overrides": map[string]interface{}{
							"communityName": "catppuccin",
							"headerBg":      "#ff0000",
						},
					},
				},
			},
			wantCommunity: "catppuccin",
			wantOverrides: true,
		},
		{
			name: "no-op when no overrides",
			input: map[string]interface{}{
				"ui": map[string]interface{}{
					"theme": map[string]interface{}{
						"name": "default",
					},
				},
			},
			wantCommunity: "",
			wantOverrides: false,
		},
		{
			name: "no-op when community already set",
			input: map[string]interface{}{
				"ui": map[string]interface{}{
					"theme": map[string]interface{}{
						"name":      "default",
						"community": "existing-theme",
						"overrides": map[string]interface{}{
							"communityName": "catppuccin",
						},
					},
				},
			},
			wantCommunity: "existing-theme",
			wantOverrides: true, // overrides untouched since community already set
		},
		{
			name:          "no-op when no ui section",
			input:         map[string]interface{}{"plugins": map[string]interface{}{}},
			wantCommunity: "",
			wantOverrides: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := migrateV1ToV2(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ui, _ := result["ui"].(map[string]interface{})
			if ui == nil {
				if tt.wantCommunity != "" {
					t.Fatalf("expected community %q but no ui section", tt.wantCommunity)
				}
				return
			}

			theme, _ := ui["theme"].(map[string]interface{})
			if theme == nil {
				if tt.wantCommunity != "" {
					t.Fatalf("expected community %q but no theme section", tt.wantCommunity)
				}
				return
			}

			community, _ := theme["community"].(string)
			if community != tt.wantCommunity {
				t.Errorf("community = %q, want %q", community, tt.wantCommunity)
			}

			_, hasOverrides := theme["overrides"]
			if hasOverrides != tt.wantOverrides {
				t.Errorf("hasOverrides = %v, want %v", hasOverrides, tt.wantOverrides)
			}
		})
	}
}

func TestRunMigrations(t *testing.T) {
	t.Run("no file is a no-op", func(t *testing.T) {
		err := RunMigrations("/nonexistent/path/config.json")
		if err != nil {
			t.Fatalf("expected no error for missing file, got: %v", err)
		}
	})

	t.Run("already at current version is a no-op", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		raw := map[string]interface{}{
			"version": float64(CurrentVersion),
			"ui":      map[string]interface{}{"showClock": true},
		}
		writeJSON(t, path, raw)

		err := RunMigrations(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// No backup should be created
		bakPath := path + ".bak"
		if _, err := os.Stat(bakPath); !os.IsNotExist(err) {
			t.Error("backup file should not be created when no migration runs")
		}
	})

	t.Run("v1 config migrates to current version", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		raw := map[string]interface{}{
			"ui": map[string]interface{}{
				"theme": map[string]interface{}{
					"name": "default",
					"overrides": map[string]interface{}{
						"communityName": "catppuccin",
					},
				},
			},
			"plugins": map[string]interface{}{},
		}
		writeJSON(t, path, raw)

		err := RunMigrations(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check backup was created
		bakPath := path + ".bak"
		if _, err := os.Stat(bakPath); os.IsNotExist(err) {
			t.Fatal("backup file should be created")
		}

		// Read migrated config
		migrated := readJSON(t, path)

		// Check version
		v, ok := migrated["version"].(float64)
		if !ok || int(v) != CurrentVersion {
			t.Errorf("version = %v, want %d", migrated["version"], CurrentVersion)
		}

		// Check community was migrated
		ui := migrated["ui"].(map[string]interface{})
		theme := ui["theme"].(map[string]interface{})
		community, _ := theme["community"].(string)
		if community != "catppuccin" {
			t.Errorf("community = %q, want %q", community, "catppuccin")
		}

		// Check overrides removed (only had communityName)
		if _, hasOverrides := theme["overrides"]; hasOverrides {
			t.Error("overrides should be removed when only communityName was present")
		}
	})

	t.Run("unknown keys preserved through migration", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		raw := map[string]interface{}{
			"ui":           map[string]interface{}{},
			"customKey":    "preserved",
			"anotherField": float64(42),
		}
		writeJSON(t, path, raw)

		err := RunMigrations(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		migrated := readJSON(t, path)

		if migrated["customKey"] != "preserved" {
			t.Errorf("customKey = %v, want %q", migrated["customKey"], "preserved")
		}
		if migrated["anotherField"] != float64(42) {
			t.Errorf("anotherField = %v, want 42", migrated["anotherField"])
		}
	})

	t.Run("idempotent — running twice produces same result", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		raw := map[string]interface{}{
			"ui": map[string]interface{}{
				"theme": map[string]interface{}{
					"name": "default",
					"overrides": map[string]interface{}{
						"communityName": "catppuccin",
					},
				},
			},
		}
		writeJSON(t, path, raw)

		// First migration
		if err := RunMigrations(path); err != nil {
			t.Fatalf("first migration failed: %v", err)
		}

		first := readJSON(t, path)

		// Second migration (should be a no-op since version is now current)
		if err := RunMigrations(path); err != nil {
			t.Fatalf("second migration failed: %v", err)
		}

		second := readJSON(t, path)

		firstBytes, _ := json.Marshal(first)
		secondBytes, _ := json.Marshal(second)
		if string(firstBytes) != string(secondBytes) {
			t.Errorf("migrations not idempotent:\nfirst:  %s\nsecond: %s", firstBytes, secondBytes)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")
		os.WriteFile(path, []byte("{invalid json"), 0644)

		err := RunMigrations(path)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("backup contains original content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.json")

		original := map[string]interface{}{
			"ui": map[string]interface{}{
				"theme": map[string]interface{}{
					"name": "default",
					"overrides": map[string]interface{}{
						"communityName": "catppuccin",
					},
				},
			},
		}
		writeJSON(t, path, original)

		originalData, _ := os.ReadFile(path)

		if err := RunMigrations(path); err != nil {
			t.Fatalf("migration failed: %v", err)
		}

		bakPath := path + ".bak"
		bakData, err := os.ReadFile(bakPath)
		if err != nil {
			t.Fatalf("failed to read backup: %v", err)
		}

		if string(bakData) != string(originalData) {
			t.Error("backup content does not match original")
		}
	})
}

// writeJSON is a test helper that writes a map as JSON to a file.
func writeJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

// readJSON is a test helper that reads a JSON file into a map.
func readJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	return m
}
