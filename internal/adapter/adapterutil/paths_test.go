package adapterutil

import (
	"runtime"
	"testing"
)

func TestResolveProjectPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantAbs bool // whether result should be absolute
	}{
		{
			name:    "empty path",
			path:    "",
			wantAbs: false,
		},
		{
			name:    "current directory",
			path:    ".",
			wantAbs: true,
		},
		{
			name:    "parent directory",
			path:    "..",
			wantAbs: true,
		},
		{
			name:    "relative path",
			path:    "some/path",
			wantAbs: true,
		},
		{
			name:    "already absolute",
			path:    "/tmp",
			wantAbs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveProjectPath(tt.path)
			if tt.wantAbs && got == "" {
				t.Errorf("ResolveProjectPath(%q) returned empty string", tt.path)
			}
		})
	}
}

func TestCWDMatchesProject(t *testing.T) {
	tests := []struct {
		name         string
		projectRoot  string
		cwd          string
		wantMatch    bool
		skipOnDarwin bool // skip on macOS due to symlink behavior differences
	}{
		{
			name:        "both empty",
			projectRoot: "",
			cwd:         "",
			wantMatch:   false,
		},
		{
			name:        "empty project root",
			projectRoot: "",
			cwd:         "/some/path",
			wantMatch:   false,
		},
		{
			name:        "empty cwd",
			projectRoot: "/some/path",
			cwd:         "",
			wantMatch:   false,
		},
		{
			name:        "cwd equals project root",
			projectRoot: "/home/user/project",
			cwd:         "/home/user/project",
			wantMatch:   true,
		},
		{
			name:        "cwd is subdirectory of project root",
			projectRoot: "/home/user/project",
			cwd:         "/home/user/project/subdir",
			wantMatch:   true,
		},
		{
			name:        "cwd is parent of project root",
			projectRoot: "/home/user/project",
			cwd:         "/home/user",
			wantMatch:   false,
		},
		{
			name:        "cwd is sibling of project root",
			projectRoot: "/home/user/project",
			cwd:         "/home/user/other",
			wantMatch:   false,
		},
		{
			name:        "cwd is unrelated path",
			projectRoot: "/home/user/project",
			cwd:         "/other/path",
			wantMatch:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnDarwin && runtime.GOOS == "darwin" {
				t.Skip("skipping on macOS due to symlink behavior differences")
			}

			got := CWDMatchesProject(tt.projectRoot, tt.cwd)
			if got != tt.wantMatch {
				t.Errorf("CWDMatchesProject(%q, %q) = %v, want %v",
					tt.projectRoot, tt.cwd, got, tt.wantMatch)
			}
		})
	}
}
