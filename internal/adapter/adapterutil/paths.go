package adapterutil

import (
	"path/filepath"
	"strings"
)

// ResolveProjectPath returns the absolute, symlink-resolved path.
// This normalizes paths by converting to absolute form and resolving symlinks.
func ResolveProjectPath(projectRoot string) string {
	if projectRoot == "" {
		return ""
	}
	abs, err := filepath.Abs(projectRoot)
	if err != nil {
		return projectRoot
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	return filepath.Clean(abs)
}

// CWDMatchesProject checks if the working directory matches the project root.
// It resolves both paths to absolute form and checks if cwd is within the project root.
func CWDMatchesProject(projectRoot, cwd string) bool {
	if projectRoot == "" || cwd == "" {
		return false
	}
	projectAbs := ResolveProjectPath(projectRoot)
	cwdAbs := ResolveProjectPath(cwd)

	if projectAbs == "" || cwdAbs == "" {
		return false
	}

	rel, err := filepath.Rel(projectAbs, cwdAbs)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return !strings.HasPrefix(rel, "..")
}
