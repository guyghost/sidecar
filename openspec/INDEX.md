# OpenSpec Index — Sidecar Architecture & Performance Improvements

**Epic**: `td-07a1d7`
**Created**: 2026-02-09
**Total Specs**: 15

## Dependency Graph

```
                    ┌──────────────────────────────────────────┐
                    │          P1 — HIGH PRIORITY              │
                    ├──────────────────────────────────────────┤
                    │                                          │
                    │  002 Immutable Styles ──┐                │
                    │                         ├──▶ 001 Model   │
                    │  003 Shared Adapter ─────┤   Decompose   │
                    │       Utils             │                │
                    │                         │                │
                    └─────────────────────────┴────────────────┘
                                    │
                    ┌───────────────▼──────────────────────────┐
                    │          P2 — MEDIUM PRIORITY            │
                    ├──────────────────────────────────────────┤
                    │                                          │
                    │  004 Unify SQLite                        │
                    │  005 Extract git/ ────▶ 007 EditorHost   │
                    │  006 Parallel Detection                  │
                    │  008 Incremental Parse (needs 003)       │
                    │  012 Keymap Type Safety                  │
                    │  013 Event Bus Improvements              │
                    │  014 Adapter Degradation                 │
                    │                                          │
                    └──────────────────────────────────────────┘
                                    │
                    ┌───────────────▼──────────────────────────┐
                    │          P3 — LOW PRIORITY               │
                    ├──────────────────────────────────────────┤
                    │                                          │
                    │  009 Config Hot Reload (needs 002, 012)  │
                    │  010 Config Schema Versioning            │
                    │  011 Perf Metrics (needs 013)            │
                    │  015 Uniform Benchmarks (needs 003)      │
                    │                                          │
                    └──────────────────────────────────────────┘
```

## Specs by Priority

### P1 — Critical (do first)

| # | Spec | TD ID | Points | Dependencies |
|---|------|-------|--------|--------------|
| 001 | [Decompose app/Model God Object](specs/001-decompose-app-model.md) | `td-b7e1a2` | 8 | 002 (recommended first) |
| 002 | [Immutable Styles System](specs/002-immutable-styles.md) | `td-396c71` | 5 | None |
| 003 | [Extract Shared Adapter Utilities](specs/003-shared-adapter-utils.md) | `td-2002fa` | 3 | None |

### P2 — Important

| # | Spec | TD ID | Points | Dependencies |
|---|------|-------|--------|--------------|
| 004 | [Unify SQLite Dependency](specs/004-unify-sqlite.md) | `td-9a9f85` | 3 | None |
| 005 | [Extract internal/git Package](specs/005-extract-git-package.md) | `td-7e3911` | 8 | None |
| 006 | [Parallelize Adapter Detection](specs/006-parallel-adapter-detection.md) | `td-1e5298` | 3 | None |
| 007 | [Reusable InlineEditorHost](specs/007-inline-editor-host.md) | `td-f35907` | 5 | None |
| 008 | [Incremental Parsing for All Adapters](specs/008-incremental-parsing.md) | `td-0da8fe` | 5 | 003 |
| 012 | [Keymap FocusContext Type Safety](specs/012-keymap-type-safety.md) | `td-31fc1f` | 2 | None |
| 013 | [Event Bus Improvements](specs/013-event-bus-improvements.md) | `td-3bf96f` | 3 | None |
| 014 | [Adapter Graceful Degradation](specs/014-adapter-graceful-degradation.md) | `td-73f37d` | 5 | None |

### P3 — Nice to Have

| # | Spec | TD ID | Points | Dependencies |
|---|------|-------|--------|--------------|
| 009 | [Config Hot Reload](specs/009-config-hot-reload.md) | `td-b7b55a` | 5 | 002, 012 |
| 010 | [Config Schema Versioning](specs/010-config-schema-versioning.md) | `td-b8cd65` | 3 | None |
| 011 | [Performance Metrics](specs/011-performance-metrics.md) | `td-9d1446` | 5 | 013 |
| 015 | [Uniform Adapter Benchmarks](specs/015-uniform-adapter-benchmarks.md) | `td-64e7a8` | 3 | 003 |

## Recommended Execution Order

```
Phase 1 (foundations):    002 → 003 → 012 → 013
Phase 2 (core refactor):  001 → 004 → 005
Phase 3 (performance):    006 → 008 → 014
Phase 4 (components):     007
Phase 5 (observability):  011 → 015
Phase 6 (config):         010 → 009
```

## Total Effort

| Priority | Specs | Story Points |
|----------|-------|-------------|
| P1 | 3 | 16 |
| P2 | 8 | 34 |
| P3 | 4 | 16 |
| **Total** | **15** | **66** |
