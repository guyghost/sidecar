# OpenSpec: Config Schema Versioning and Migration

**ID**: `td-b8cd65`
**Epic**: `td-07a1d7` — Sidecar Architecture & Performance Improvements
**Priority**: P3
**Type**: Feature
**Effort**: 3 story points

## Problem Statement

`config.json` has no schema version. As the config format evolves, there is no mechanism to detect outdated configs or automatically migrate them. The current approach relies on ad-hoc legacy field detection (e.g., `communityName` migration) which does not scale.

## Objective

Add a `version` field to config.json with a forward-compatible automatic migration system.

## Constraints

- Existing configs without `version` field must be auto-detected as v1
- Migration must be non-destructive (backup before modifying)
- Migration must be idempotent (running twice produces same result)
- Config must remain human-editable (no binary format)

## Technical Design

### Config Version Format

```json
{
  "version": 2,
  "theme": { ... },
  "plugins": { ... }
}
```

Simple integer versioning. Current config format = version 1 (implicit).

### Migration System

```go
// internal/config/migrate.go

type Migration struct {
    FromVersion int
    ToVersion   int
    Migrate     func(raw map[string]interface{}) (map[string]interface{}, error)
    Description string
}

var migrations = []Migration{
    {
        FromVersion: 1,
        ToVersion:   2,
        Migrate:     migrateV1ToV2,
        Description: "Move communityName to theme.community, add version field",
    },
}

// RunMigrations applies all necessary migrations sequentially.
func RunMigrations(configPath string) error {
    raw := loadRawJSON(configPath)
    version := detectVersion(raw) // 0 or missing → 1
    
    for _, m := range migrations {
        if version == m.FromVersion {
            backup(configPath) // config.json.bak
            raw, _ = m.Migrate(raw)
            version = m.ToVersion
        }
    }
    
    raw["version"] = version
    saveRawJSON(configPath, raw)
    return nil
}
```

### Version Detection

```go
func detectVersion(raw map[string]interface{}) int {
    if v, ok := raw["version"].(float64); ok {
        return int(v)
    }
    return 1 // implicit v1
}
```

### Backward Compatibility

- Unknown keys are always preserved (already implemented in `saver.go`)
- Old sidecar versions ignore the `version` field (it's just another unknown key)
- Migrations only run forward, never backward

## Acceptance Criteria

- [ ] New configs include `"version": N` (where N is current version)
- [ ] Existing configs without version are treated as v1
- [ ] Migration system applies sequential migrations (v1→v2→v3)
- [ ] Backup is created before any migration (`config.json.bak`)
- [ ] Migrations are idempotent
- [ ] Unknown keys are preserved through migration
- [ ] `go test ./internal/config/...` includes migration tests
- [ ] Saver includes version in output

## Dependencies

- None

## Risks

- **Low**: Users manually editing version field to invalid value
- **Mitigation**: Validate version range, clamp to max known version

## Scenarios

### Scenario: Fresh install
```
Given no config.json exists
When sidecar creates default config
Then config.json includes "version": N (current latest)
```

### Scenario: Legacy config upgrade
```
Given a config.json without version field and using "communityName"
When sidecar starts
Then config is detected as v1
And migration v1→v2 runs (moving communityName to theme.community)
And config.json now has "version": 2
And config.json.bak contains the original
```

### Scenario: Already up-to-date config
```
Given a config.json with "version": 3 (current latest)
When sidecar starts
Then no migrations run
And config is loaded normally
```
