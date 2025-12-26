# Agents: Mandatory: use td usage --new-session to see open work

If you don't have an existing task for the work you're about to start, create one first.

## Build & Versioning

```bash
# Build
go build ./...

# Run tests
go test ./...

# Install with version (use semantic versioning)
go install -ldflags "-X main.Version=v0.1.0" ./cmd/sidecar

# Tag a release
git tag v0.1.0 -m "Release message"
```

Version is set via ldflags at build time. Without it, sidecar shows git revision info.
