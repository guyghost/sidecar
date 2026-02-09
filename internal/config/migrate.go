package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// CurrentVersion is the latest config schema version.
const CurrentVersion = 2

// Migration describes a single schema migration step.
type Migration struct {
	FromVersion int
	ToVersion   int
	Migrate     func(raw map[string]interface{}) (map[string]interface{}, error)
	Description string
}

// migrations is the ordered registry of all schema migrations.
// Each migration transforms from one version to the next.
var migrations = []Migration{
	{
		FromVersion: 1,
		ToVersion:   2,
		Migrate:     migrateV1ToV2,
		Description: "Move communityName from overrides to theme.community, add version field",
	},
}

// detectVersion returns the schema version from a raw config map.
// Configs without a version field are treated as v1.
func detectVersion(raw map[string]interface{}) int {
	v, ok := raw["version"]
	if !ok {
		return 1
	}
	// JSON numbers unmarshal as float64
	if f, ok := v.(float64); ok {
		n := int(f)
		if n < 1 {
			return 1
		}
		if n > CurrentVersion {
			return CurrentVersion
		}
		return n
	}
	return 1
}

// RunMigrations reads the config file, applies any necessary migrations,
// and writes the result back. A backup (.bak) is created before the first
// migration modifies the file. If the file does not exist or is already at
// the current version, this is a no-op.
func RunMigrations(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no file → nothing to migrate
		}
		return fmt.Errorf("read config for migration: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse config for migration: %w", err)
	}

	version := detectVersion(raw)
	if version >= CurrentVersion {
		return nil // already up to date
	}

	// Create backup before any migration
	bakPath := configPath + ".bak"
	if err := os.WriteFile(bakPath, data, 0644); err != nil {
		return fmt.Errorf("create config backup: %w", err)
	}
	slog.Info("config: created backup before migration", "backup", bakPath)

	// Apply migrations sequentially
	for _, m := range migrations {
		if version == m.FromVersion {
			slog.Info("config: running migration",
				"from", m.FromVersion,
				"to", m.ToVersion,
				"description", m.Description,
			)
			raw, err = m.Migrate(raw)
			if err != nil {
				return fmt.Errorf("migration v%d→v%d: %w", m.FromVersion, m.ToVersion, err)
			}
			version = m.ToVersion
		}
	}

	// Stamp final version
	raw["version"] = float64(version)

	// Write migrated config back
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal migrated config: %w", err)
	}
	if err := os.WriteFile(configPath, out, 0644); err != nil {
		return fmt.Errorf("write migrated config: %w", err)
	}

	slog.Info("config: migration complete", "version", version)
	return nil
}

// migrateV1ToV2 moves communityName from ui.theme.overrides to ui.theme.community.
// This replaces the ad-hoc migration that was previously inlined in mergeConfig().
func migrateV1ToV2(raw map[string]interface{}) (map[string]interface{}, error) {
	ui, ok := raw["ui"].(map[string]interface{})
	if !ok {
		return raw, nil // no ui section, nothing to migrate
	}

	theme, ok := ui["theme"].(map[string]interface{})
	if !ok {
		return raw, nil // no theme section
	}

	overrides, ok := theme["overrides"].(map[string]interface{})
	if !ok {
		return raw, nil // no overrides
	}

	communityName, ok := overrides["communityName"]
	if !ok {
		return raw, nil // no communityName in overrides
	}

	nameStr, ok := communityName.(string)
	if !ok || nameStr == "" {
		return raw, nil // not a valid string
	}

	// Only migrate if community is not already set
	existing, _ := theme["community"].(string)
	if existing != "" {
		return raw, nil // community already set, don't overwrite
	}

	// Move communityName → community
	theme["community"] = nameStr
	delete(overrides, "communityName")

	// If overrides is now empty, remove it
	if len(overrides) == 0 {
		delete(theme, "overrides")
	}

	return raw, nil
}
