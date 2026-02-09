# Context: Architecture & Performance Improvements

## Objective
Implement all 15 OpenSpec improvements across 6 phases, covering: immutable styles, shared adapter utilities, keymap type safety, event bus improvements, Model decomposition, SQLite unification, git package extraction, parallel adapter detection, incremental parsing, adapter degradation, inline editor host, performance metrics, benchmarks, config versioning, and config hot reload.

## Constraints
- Platform: Terminal/TUI (Bubble Tea + lipgloss)
- Mode: Ralph (autonomous loop, max 10 iterations)
- Zero behavioral regressions — all tests must pass
- Incremental delivery — each spec is independently deployable

## Execution Plan
```
Phase 1 (foundations):    002 → 003 → 012 → 013
Phase 2 (core refactor):  001 → 004 → 005
Phase 3 (performance):    006 → 008 → 014
Phase 4 (components):     007
Phase 5 (observability):  011 → 015
Phase 6 (config):         010 → 009
```

## Technical Decisions
| Decision | Justification | Agent |
|----------|---------------|-------|
| Ralph mode enabled | User requested autonomous implementation of all phases | @orchestrator |
| Skip DESIGN phase | Pure Go refactoring, no visual design needed | @orchestrator |
| Phase 1 first | Styles + adapter utils are foundations for later phases | @orchestrator |

## Artifacts Produced
| File | Agent | Status |
|------|-------|--------|

## Inter-Agent Notes
