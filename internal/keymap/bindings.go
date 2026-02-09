package keymap

// DefaultBindings returns the default keymap.
func DefaultBindings() []Binding {
	return []Binding{
		// Global context
		{Key: "q", Command: "quit", Context: ContextGlobal},
		{Key: "?", Command: "toggle-palette", Context: ContextGlobal},
		{Key: "!", Command: "toggle-diagnostics", Context: ContextGlobal},
		{Key: "`", Command: "next-plugin", Context: ContextGlobal},
		{Key: "~", Command: "prev-plugin", Context: ContextGlobal},
		{Key: "@", Command: "switch-project", Context: ContextGlobal},
		{Key: "1", Command: "focus-plugin-1", Context: ContextGlobal},
		{Key: "2", Command: "focus-plugin-2", Context: ContextGlobal},
		{Key: "3", Command: "focus-plugin-3", Context: ContextGlobal},
		{Key: "4", Command: "focus-plugin-4", Context: ContextGlobal},
		{Key: "5", Command: "focus-plugin-5", Context: ContextGlobal},
		{Key: "6", Command: "focus-plugin-6", Context: ContextGlobal},
		{Key: "7", Command: "focus-plugin-7", Context: ContextGlobal},
		{Key: "8", Command: "focus-plugin-8", Context: ContextGlobal},
		{Key: "9", Command: "focus-plugin-9", Context: ContextGlobal},

		// Navigation (Global defaults)
		{Key: "j", Command: "cursor-down", Context: ContextGlobal},
		{Key: "k", Command: "cursor-up", Context: ContextGlobal},
		{Key: "down", Command: "cursor-down", Context: ContextGlobal},
		{Key: "up", Command: "cursor-up", Context: ContextGlobal},
		{Key: "ctrl+n", Command: "cursor-down", Context: ContextGlobal},
		{Key: "ctrl+p", Command: "cursor-up", Context: ContextGlobal},
		{Key: "g g", Command: "cursor-top", Context: ContextGlobal},
		{Key: "G", Command: "cursor-bottom", Context: ContextGlobal},
		{Key: "enter", Command: "select", Context: ContextGlobal},
		{Key: "esc", Command: "back", Context: ContextGlobal},

		// Project switcher context
		{Key: "@", Command: "toggle", Context: ContextProjectSwitcher},
		{Key: "esc", Command: "close", Context: ContextProjectSwitcher},
		{Key: "enter", Command: "select", Context: ContextProjectSwitcher},
		{Key: "down", Command: "cursor-down", Context: ContextProjectSwitcher},
		{Key: "up", Command: "cursor-up", Context: ContextProjectSwitcher},
		{Key: "ctrl+n", Command: "cursor-down", Context: ContextProjectSwitcher},
		{Key: "ctrl+p", Command: "cursor-up", Context: ContextProjectSwitcher},

		// Git status context
		{Key: "i", Command: "init-repo", Context: ContextGitNoRepo},
		{Key: "enter", Command: "init-repo", Context: ContextGitNoRepo},
		{Key: "r", Command: "refresh", Context: ContextGitNoRepo},

		{Key: "j", Command: "cursor-down", Context: ContextGitStatus},
		{Key: "k", Command: "cursor-up", Context: ContextGitStatus},
		{Key: "tab", Command: "switch-pane", Context: ContextGitStatus},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextGitStatus},
		{Key: "s", Command: "stage-file", Context: ContextGitStatus},
		{Key: "u", Command: "unstage-file", Context: ContextGitStatus},
		{Key: "S", Command: "stage-all", Context: ContextGitStatus},
		{Key: "U", Command: "unstage-all", Context: ContextGitStatus},
		{Key: "c", Command: "commit", Context: ContextGitStatus},
		{Key: "A", Command: "amend", Context: ContextGitStatus},
		{Key: "d", Command: "show-diff", Context: ContextGitStatus},
		{Key: "enter", Command: "show-diff", Context: ContextGitStatus},
		{Key: "r", Command: "refresh", Context: ContextGitStatus},
		{Key: "h", Command: "show-history", Context: ContextGitStatus},
		{Key: "P", Command: "push", Context: ContextGitStatus},
		{Key: "f", Command: "fetch", Context: ContextGitStatus},
		{Key: "L", Command: "pull", Context: ContextGitStatus},
		{Key: "b", Command: "branch-picker", Context: ContextGitStatus},
		{Key: "z", Command: "stash", Context: ContextGitStatus},
		{Key: "Z", Command: "stash-pop", Context: ContextGitStatus},
		{Key: "ctrl+z", Command: "stash-apply", Context: ContextGitStatus},
		{Key: "O", Command: "open-in-file-browser", Context: ContextGitStatus},
		{Key: "o", Command: "open-in-github", Context: ContextGitStatus},
		{Key: "y", Command: "yank-file", Context: ContextGitStatus},
		{Key: "Y", Command: "yank-path", Context: ContextGitStatus},
		{Key: "D", Command: "discard-changes", Context: ContextGitStatus},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextGitStatus},

		// Git status commits context (sidebar)
		{Key: "j", Command: "cursor-down", Context: ContextGitStatusCommits},
		{Key: "k", Command: "cursor-up", Context: ContextGitStatusCommits},
		{Key: "enter", Command: "view-commit", Context: ContextGitStatusCommits},
		{Key: "d", Command: "view-commit", Context: ContextGitStatusCommits},
		{Key: "h", Command: "show-history", Context: ContextGitStatusCommits},
		{Key: "y", Command: "yank-commit", Context: ContextGitStatusCommits},
		{Key: "Y", Command: "yank-id", Context: ContextGitStatusCommits},
		{Key: "/", Command: "search-history", Context: ContextGitStatusCommits},
		{Key: "f", Command: "filter-author", Context: ContextGitStatusCommits},
		{Key: "p", Command: "filter-path", Context: ContextGitStatusCommits},
		{Key: "F", Command: "clear-filter", Context: ContextGitStatusCommits},
		{Key: "n", Command: "next-match", Context: ContextGitStatusCommits},
		{Key: "N", Command: "prev-match", Context: ContextGitStatusCommits},
		{Key: "o", Command: "open-in-github", Context: ContextGitStatusCommits},
		{Key: "v", Command: "toggle-graph", Context: ContextGitStatusCommits},
		{Key: "P", Command: "push", Context: ContextGitStatusCommits},
		{Key: "L", Command: "pull", Context: ContextGitStatusCommits},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextGitStatusCommits},

		// Git history search modal context
		{Key: "enter", Command: "select", Context: ContextGitHistorySearch},
		{Key: "esc", Command: "cancel", Context: ContextGitHistorySearch},
		{Key: "j", Command: "navigate", Context: ContextGitHistorySearch},
		{Key: "k", Command: "navigate", Context: ContextGitHistorySearch},
		{Key: "down", Command: "navigate", Context: ContextGitHistorySearch},
		{Key: "up", Command: "navigate", Context: ContextGitHistorySearch},
		{Key: "alt+r", Command: "toggle-regex", Context: ContextGitHistorySearch},
		{Key: "alt+c", Command: "toggle-case", Context: ContextGitHistorySearch},

		// Git path filter modal context
		{Key: "enter", Command: "apply-filter", Context: ContextGitPathFilter},
		{Key: "esc", Command: "cancel", Context: ContextGitPathFilter},

		// Git status diff context (inline)
		{Key: "j", Command: "scroll-down", Context: ContextGitStatusDiff},
		{Key: "k", Command: "scroll-up", Context: ContextGitStatusDiff},
		{Key: "ctrl+d", Command: "page-down", Context: ContextGitStatusDiff},
		{Key: "ctrl+u", Command: "page-up", Context: ContextGitStatusDiff},
		{Key: "enter", Command: "full-diff", Context: ContextGitStatusDiff},
		{Key: "s", Command: "stage-file", Context: ContextGitStatusDiff},
		{Key: "u", Command: "unstage-file", Context: ContextGitStatusDiff},
		{Key: "v", Command: "toggle-diff-view", Context: ContextGitStatusDiff},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextGitStatusDiff},
		{Key: "w", Command: "toggle-wrap", Context: ContextGitStatusDiff},

		// Git commit preview context
		{Key: "j", Command: "scroll-down", Context: ContextGitCommitPreview},
		{Key: "k", Command: "scroll-up", Context: ContextGitCommitPreview},
		{Key: "d", Command: "view-diff", Context: ContextGitCommitPreview},
		{Key: "esc", Command: "back", Context: ContextGitCommitPreview},
		{Key: "y", Command: "yank-commit", Context: ContextGitCommitPreview},
		{Key: "Y", Command: "yank-id", Context: ContextGitCommitPreview},
		{Key: "o", Command: "open-in-github", Context: ContextGitCommitPreview},
		{Key: "b", Command: "open-in-file-browser", Context: ContextGitCommitPreview},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextGitCommitPreview},

		// Git diff context (full screen)
		{Key: "esc", Command: "close-diff", Context: ContextGitDiff},
		{Key: "q", Command: "close-diff", Context: ContextGitDiff},
		{Key: "j", Command: "scroll-down", Context: ContextGitDiff},
		{Key: "k", Command: "scroll-up", Context: ContextGitDiff},
		{Key: "down", Command: "scroll-down", Context: ContextGitDiff},
		{Key: "up", Command: "scroll-up", Context: ContextGitDiff},
		{Key: "ctrl+d", Command: "page-down", Context: ContextGitDiff},
		{Key: "ctrl+u", Command: "page-up", Context: ContextGitDiff},
		{Key: "s", Command: "stage-file", Context: ContextGitDiff},
		{Key: "u", Command: "unstage-file", Context: ContextGitDiff},
		{Key: "[", Command: "prev-file", Context: ContextGitDiff},
		{Key: "]", Command: "next-file", Context: ContextGitDiff},
		{Key: "y", Command: "yank-diff", Context: ContextGitDiff},
		{Key: "c", Command: "commit", Context: ContextGitDiff},
		{Key: "v", Command: "toggle-diff-view", Context: ContextGitDiff},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextGitDiff},
		{Key: "w", Command: "toggle-wrap", Context: ContextGitDiff},

		// Git push menu context
		{Key: "p", Command: "push", Context: ContextGitPushMenu},
		{Key: "f", Command: "force-push", Context: ContextGitPushMenu},
		{Key: "u", Command: "push-upstream", Context: ContextGitPushMenu},
		{Key: "esc", Command: "cancel", Context: ContextGitPushMenu},

		// Git pull menu context
		{Key: "p", Command: "pull-merge", Context: ContextGitPullMenu},
		{Key: "r", Command: "pull-rebase", Context: ContextGitPullMenu},
		{Key: "f", Command: "pull-ff-only", Context: ContextGitPullMenu},
		{Key: "a", Command: "pull-autostash", Context: ContextGitPullMenu},
		{Key: "esc", Command: "cancel", Context: ContextGitPullMenu},

		// Issue preview context
		// Issue input modal context
		{Key: "ctrl+x", Command: "toggle-closed", Context: ContextIssueInput},

		{Key: "o", Command: "open-in-td", Context: ContextIssuePreview},
		{Key: "b", Command: "issue-back", Context: ContextIssuePreview},
		{Key: "y", Command: "yank-issue", Context: ContextIssuePreview},
		{Key: "Y", Command: "yank-issue-key", Context: ContextIssuePreview},
		{Key: "esc", Command: "close", Context: ContextIssuePreview},

		// Git error modal context
		{Key: "L", Command: "pull-from-error", Context: ContextGitError},
		{Key: "y", Command: "yank-error", Context: ContextGitError},
		{Key: "esc", Command: "dismiss", Context: ContextGitError},

		// Git pull conflict context
		{Key: "a", Command: "abort-pull", Context: ContextGitPullConflict},
		{Key: "esc", Command: "dismiss", Context: ContextGitPullConflict},

		// Git stash pop context
		{Key: "y", Command: "confirm-pop", Context: ContextGitStashPop},
		{Key: "esc", Command: "dismiss", Context: ContextGitStashPop},

		// Git commit context
		{Key: "ctrl+s", Command: "execute-commit", Context: ContextGitCommit},
		{Key: "ctrl+enter", Command: "execute-commit", Context: ContextGitCommit},
		{Key: "esc", Command: "cancel", Context: ContextGitCommit},

		// Git history context
		{Key: "esc", Command: "close-history", Context: ContextGitHistory},
		{Key: "q", Command: "close-history", Context: ContextGitHistory},
		{Key: "enter", Command: "view-commit", Context: ContextGitHistory},

		// Git commit detail context
		{Key: "esc", Command: "close-detail", Context: ContextGitCommitDetail},
		{Key: "q", Command: "close-detail", Context: ContextGitCommitDetail},

		// Conversations sidebar context (two-pane mode, left pane focused)
		{Key: "tab", Command: "switch-pane", Context: ContextConversationsSidebar},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextConversationsSidebar},
		{Key: "a", Command: "new-session", Context: ContextConversationsSidebar},
		{Key: "d", Command: "delete-session", Context: ContextConversationsSidebar},
		{Key: "r", Command: "rename-session", Context: ContextConversationsSidebar},
		{Key: "e", Command: "export-session", Context: ContextConversationsSidebar},
		{Key: "c", Command: "copy-session", Context: ContextConversationsSidebar},
		{Key: "f", Command: "filter", Context: ContextConversationsSidebar},
		{Key: "/", Command: "search", Context: ContextConversationsSidebar},
		{Key: "s", Command: "toggle-star", Context: ContextConversationsSidebar},
		{Key: "A", Command: "show-analytics", Context: ContextConversationsSidebar},
		{Key: "l", Command: "focus-right", Context: ContextConversationsSidebar},
		{Key: "right", Command: "focus-right", Context: ContextConversationsSidebar},
		{Key: "v", Command: "toggle-view", Context: ContextConversationsSidebar},
		{Key: "enter", Command: "select-session", Context: ContextConversationsSidebar},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextConversationsSidebar},
		{Key: "y", Command: "yank-details", Context: ContextConversationsSidebar},
		{Key: "Y", Command: "yank-resume", Context: ContextConversationsSidebar},
		{Key: "R", Command: "resume-in-workspace", Context: ContextConversationsSidebar},

		// Conversations main context (two-pane mode, right pane focused)
		{Key: "tab", Command: "switch-pane", Context: ContextConversationsMain},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextConversationsMain},
		{Key: "esc", Command: "back", Context: ContextConversationsMain},
		{Key: "j", Command: "scroll", Context: ContextConversationsMain},
		{Key: "k", Command: "scroll", Context: ContextConversationsMain},
		{Key: "g", Command: "cursor-top", Context: ContextConversationsMain},
		{Key: "G", Command: "cursor-bottom", Context: ContextConversationsMain},
		{Key: "h", Command: "focus-left", Context: ContextConversationsMain},
		{Key: "left", Command: "focus-left", Context: ContextConversationsMain},
		{Key: "v", Command: "toggle-view", Context: ContextConversationsMain},
		{Key: "e", Command: "expand", Context: ContextConversationsMain},
		{Key: "enter", Command: "detail", Context: ContextConversationsMain},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextConversationsMain},
		{Key: "y", Command: "yank-details", Context: ContextConversationsMain},
		{Key: "Y", Command: "yank-resume", Context: ContextConversationsMain},
		{Key: "R", Command: "resume-in-workspace", Context: ContextConversationsMain},

		// File browser tree context
		{Key: "tab", Command: "switch-pane", Context: ContextFileBrowserTree},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextFileBrowserTree},
		{Key: "/", Command: "search", Context: ContextFileBrowserTree},
		{Key: "ctrl+p", Command: "quick-open", Context: ContextFileBrowserTree},
		{Key: "f", Command: "project-search", Context: ContextFileBrowserTree},
		{Key: "t", Command: "new-tab", Context: ContextFileBrowserTree},
		{Key: "[", Command: "prev-tab", Context: ContextFileBrowserTree},
		{Key: "]", Command: "next-tab", Context: ContextFileBrowserTree},
		{Key: "x", Command: "close-tab", Context: ContextFileBrowserTree},
		{Key: "a", Command: "create-file", Context: ContextFileBrowserTree},
		{Key: "A", Command: "create-dir", Context: ContextFileBrowserTree},
		{Key: "D", Command: "delete", Context: ContextFileBrowserTree},
		{Key: "y", Command: "yank", Context: ContextFileBrowserTree},
		{Key: "Y", Command: "copy-path", Context: ContextFileBrowserTree},
		{Key: "p", Command: "paste", Context: ContextFileBrowserTree},
		{Key: "s", Command: "sort", Context: ContextFileBrowserTree},
		{Key: "r", Command: "refresh", Context: ContextFileBrowserTree},
		{Key: "m", Command: "move", Context: ContextFileBrowserTree},
		{Key: "R", Command: "rename", Context: ContextFileBrowserTree},
		{Key: "ctrl+r", Command: "reveal", Context: ContextFileBrowserTree},
		{Key: "I", Command: "info", Context: ContextFileBrowserTree},
		{Key: "e", Command: "edit", Context: ContextFileBrowserTree},
		{Key: "E", Command: "edit-external", Context: ContextFileBrowserTree},
		{Key: "B", Command: "blame", Context: ContextFileBrowserTree},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextFileBrowserTree},
		{Key: "H", Command: "toggle-ignored", Context: ContextFileBrowserTree},

		// File browser preview context
		{Key: "tab", Command: "switch-pane", Context: ContextFileBrowserPreview},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextFileBrowserPreview},
		{Key: "/", Command: "search-content", Context: ContextFileBrowserPreview},
		{Key: "ctrl+p", Command: "quick-open", Context: ContextFileBrowserPreview},
		{Key: "f", Command: "project-search", Context: ContextFileBrowserPreview},
		{Key: "[", Command: "prev-tab", Context: ContextFileBrowserPreview},
		{Key: "]", Command: "next-tab", Context: ContextFileBrowserPreview},
		{Key: "x", Command: "close-tab", Context: ContextFileBrowserPreview},
		{Key: "r", Command: "refresh", Context: ContextFileBrowserPreview},
		{Key: "R", Command: "rename", Context: ContextFileBrowserPreview},
		{Key: "ctrl+r", Command: "reveal", Context: ContextFileBrowserPreview},
		{Key: "I", Command: "info", Context: ContextFileBrowserPreview},
		{Key: "e", Command: "edit", Context: ContextFileBrowserPreview},
		{Key: "E", Command: "edit-external", Context: ContextFileBrowserPreview},
		{Key: "B", Command: "blame", Context: ContextFileBrowserPreview},
		{Key: "m", Command: "toggle-markdown", Context: ContextFileBrowserPreview},
		{Key: "esc", Command: "back", Context: ContextFileBrowserPreview},
		{Key: "h", Command: "back", Context: ContextFileBrowserPreview},
		{Key: "y", Command: "yank-contents", Context: ContextFileBrowserPreview},
		{Key: "Y", Command: "yank-path", Context: ContextFileBrowserPreview},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextFileBrowserPreview},
		{Key: "w", Command: "toggle-wrap", Context: ContextFileBrowserPreview},

		// File browser tree search context
		{Key: "esc", Command: "cancel", Context: ContextFileBrowserSearch},
		{Key: "enter", Command: "confirm", Context: ContextFileBrowserSearch},
		{Key: "n", Command: "next-match", Context: ContextFileBrowserSearch},
		{Key: "N", Command: "prev-match", Context: ContextFileBrowserSearch},

		// File browser content search context
		{Key: "esc", Command: "cancel", Context: ContextFileBrowserContentSearch},
		{Key: "enter", Command: "confirm", Context: ContextFileBrowserContentSearch},
		{Key: "n", Command: "next-match", Context: ContextFileBrowserContentSearch},
		{Key: "N", Command: "prev-match", Context: ContextFileBrowserContentSearch},

		// File browser quick open context
		{Key: "esc", Command: "cancel", Context: ContextFileBrowserQuickOpen},
		{Key: "enter", Command: "select", Context: ContextFileBrowserQuickOpen},
		{Key: "up", Command: "cursor-up", Context: ContextFileBrowserQuickOpen},
		{Key: "down", Command: "cursor-down", Context: ContextFileBrowserQuickOpen},
		{Key: "ctrl+n", Command: "cursor-down", Context: ContextFileBrowserQuickOpen},
		{Key: "ctrl+p", Command: "cursor-up", Context: ContextFileBrowserQuickOpen},

		// File browser project search context
		{Key: "esc", Command: "cancel", Context: ContextFileBrowserProjectSearch},
		{Key: "enter", Command: "select", Context: ContextFileBrowserProjectSearch},
		{Key: "up", Command: "cursor-up", Context: ContextFileBrowserProjectSearch},
		{Key: "down", Command: "cursor-down", Context: ContextFileBrowserProjectSearch},
		{Key: "ctrl+n", Command: "cursor-down", Context: ContextFileBrowserProjectSearch},
		{Key: "ctrl+p", Command: "cursor-up", Context: ContextFileBrowserProjectSearch},
		{Key: "tab", Command: "toggle", Context: ContextFileBrowserProjectSearch},
		{Key: "alt+r", Command: "toggle-regex", Context: ContextFileBrowserProjectSearch},
		{Key: "alt+c", Command: "toggle-case", Context: ContextFileBrowserProjectSearch},
		{Key: "alt+w", Command: "toggle-word", Context: ContextFileBrowserProjectSearch},
		{Key: "ctrl+g", Command: "cursor-top", Context: ContextFileBrowserProjectSearch},
		{Key: "ctrl+e", Command: "open-in-editor", Context: ContextFileBrowserProjectSearch},
		{Key: "ctrl+d", Command: "page-down", Context: ContextFileBrowserProjectSearch},
		{Key: "ctrl+u", Command: "page-up", Context: ContextFileBrowserProjectSearch},

		// File browser file operation context
		{Key: "esc", Command: "cancel", Context: ContextFileBrowserFileOp},
		{Key: "enter", Command: "confirm", Context: ContextFileBrowserFileOp},
		{Key: "tab", Command: "next-button", Context: ContextFileBrowserFileOp},
		{Key: "shift+tab", Command: "prev-button", Context: ContextFileBrowserFileOp},

		// File browser line jump context
		{Key: "esc", Command: "cancel", Context: ContextFileBrowserLineJump},
		{Key: "enter", Command: "confirm", Context: ContextFileBrowserLineJump},

		// Worktree context
		{Key: "n", Command: "new-workspace", Context: ContextWorkspaceList},
		{Key: "v", Command: "toggle-view", Context: ContextWorkspaceList},
		{Key: "r", Command: "refresh", Context: ContextWorkspaceList},
		{Key: "D", Command: "delete-workspace", Context: ContextWorkspaceList},
		{Key: "d", Command: "show-diff", Context: ContextWorkspaceList},
		{Key: "p", Command: "push", Context: ContextWorkspaceList},
		{Key: "m", Command: "merge-workflow", Context: ContextWorkspaceList},
		{Key: "T", Command: "link-task", Context: ContextWorkspaceList},
		{Key: "s", Command: "start-agent", Context: ContextWorkspaceList},
		{Key: "E", Command: "interactive", Context: ContextWorkspaceList},
		{Key: "t", Command: "attach", Context: ContextWorkspaceList},
		{Key: "S", Command: "stop-agent", Context: ContextWorkspaceList},
		{Key: "y", Command: "approve", Context: ContextWorkspaceList},
		{Key: "Y", Command: "approve-all", Context: ContextWorkspaceList},
		{Key: "N", Command: "reject", Context: ContextWorkspaceList},
		{Key: "K", Command: "kill-shell", Context: ContextWorkspaceList},
		{Key: "O", Command: "open-in-git", Context: ContextWorkspaceList},
		{Key: "l", Command: "focus-right", Context: ContextWorkspaceList},
		{Key: "right", Command: "focus-right", Context: ContextWorkspaceList},
		{Key: "tab", Command: "switch-pane", Context: ContextWorkspaceList},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextWorkspaceList},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextWorkspaceList},
		{Key: "[", Command: "prev-tab", Context: ContextWorkspaceList},
		{Key: "]", Command: "next-tab", Context: ContextWorkspaceList},
		{Key: "F", Command: "fetch-pr", Context: ContextWorkspaceList},

		// Workspace fetch PR context
		{Key: "esc", Command: "cancel", Context: ContextWorkspaceFetchPR},
		{Key: "enter", Command: "fetch", Context: ContextWorkspaceFetchPR},

		// Workspace preview context
		{Key: "h", Command: "focus-left", Context: ContextWorkspacePreview},
		{Key: "left", Command: "focus-left", Context: ContextWorkspacePreview},
		{Key: "esc", Command: "focus-left", Context: ContextWorkspacePreview},
		{Key: "s", Command: "start-agent", Context: ContextWorkspacePreview},
		{Key: "S", Command: "stop-agent", Context: ContextWorkspacePreview},
		{Key: "y", Command: "approve", Context: ContextWorkspacePreview},
		{Key: "Y", Command: "approve-all", Context: ContextWorkspacePreview},
		{Key: "N", Command: "reject", Context: ContextWorkspacePreview},
		{Key: "v", Command: "toggle-diff-view", Context: ContextWorkspacePreview},
		{Key: "0", Command: "reset-scroll", Context: ContextWorkspacePreview},
		{Key: "tab", Command: "switch-pane", Context: ContextWorkspacePreview},
		{Key: "shift+tab", Command: "switch-pane", Context: ContextWorkspacePreview},
		{Key: "\\", Command: "toggle-sidebar", Context: ContextWorkspacePreview},
		{Key: "[", Command: "prev-tab", Context: ContextWorkspacePreview},
		{Key: "]", Command: "next-tab", Context: ContextWorkspacePreview},
		{Key: "j", Command: "scroll-down", Context: ContextWorkspacePreview},
		{Key: "k", Command: "scroll-up", Context: ContextWorkspacePreview},
		{Key: "ctrl+d", Command: "page-down", Context: ContextWorkspacePreview},
		{Key: "ctrl+u", Command: "page-up", Context: ContextWorkspacePreview},

		// Workspace merge error context
		{Key: "esc", Command: "dismiss-merge-error", Context: ContextWorkspaceMergeError},
		{Key: "y", Command: "yank-merge-error", Context: ContextWorkspaceMergeError},

		// Workspace interactive context bindings are registered dynamically
		// by the workspace plugin Init() to reflect configured keys.

		// Notes list context
		{Key: "j", Command: "cursor-down", Context: ContextNotesList},
		{Key: "k", Command: "cursor-up", Context: ContextNotesList},
		{Key: "down", Command: "cursor-down", Context: ContextNotesList},
		{Key: "up", Command: "cursor-up", Context: ContextNotesList},
		{Key: "G", Command: "cursor-bottom", Context: ContextNotesList},
		{Key: "n", Command: "new-note", Context: ContextNotesList},
		{Key: "X", Command: "delete-note", Context: ContextNotesList},
		{Key: "x", Command: "show-deleted", Context: ContextNotesList},
		{Key: "p", Command: "toggle-pin", Context: ContextNotesList},
		{Key: "A", Command: "archive-note", Context: ContextNotesList},
		{Key: "a", Command: "show-archived", Context: ContextNotesList},
		{Key: "u", Command: "undo", Context: ContextNotesList},
		{Key: "r", Command: "refresh", Context: ContextNotesList},
		{Key: "enter", Command: "edit-note", Context: ContextNotesList},
		{Key: "/", Command: "search", Context: ContextNotesList},
		{Key: "T", Command: "to-task", Context: ContextNotesList},
		{Key: "I", Command: "show-info", Context: ContextNotesList},
		{Key: "y", Command: "yank-content", Context: ContextNotesList},
		{Key: "Y", Command: "yank-title", Context: ContextNotesList},
		{Key: "esc", Command: "back-to-active", Context: ContextNotesList},
		{Key: "e", Command: "vim-edit", Context: ContextNotesList},
		{Key: "E", Command: "external-editor", Context: ContextNotesList},

		// Notes info modal context
		{Key: "esc", Command: "close", Context: ContextNotesInfo},
		{Key: "enter", Command: "close", Context: ContextNotesInfo},

		// Notes search context
		{Key: "esc", Command: "cancel", Context: ContextNotesSearch},
		{Key: "enter", Command: "select", Context: ContextNotesSearch},
		{Key: "down", Command: "cursor-down", Context: ContextNotesSearch},
		{Key: "up", Command: "cursor-up", Context: ContextNotesSearch},
		{Key: "ctrl+n", Command: "cursor-down", Context: ContextNotesSearch},
		{Key: "ctrl+p", Command: "cursor-up", Context: ContextNotesSearch},

		// Notes preview context (read-only view)
		{Key: "alt+c", Command: "copy-note", Context: ContextNotesPreview},
		{Key: "e", Command: "vim-edit", Context: ContextNotesPreview},
		{Key: "E", Command: "external-editor", Context: ContextNotesPreview},

		// Notes editor context
		{Key: "tab", Command: "switch-pane", Context: ContextNotesEditor},
		{Key: "esc", Command: "back", Context: ContextNotesEditor},
		{Key: "ctrl+s", Command: "save", Context: ContextNotesEditor},
		{Key: "E", Command: "external-editor", Context: ContextNotesEditor},
		{Key: "alt+c", Command: "copy-note", Context: ContextNotesEditor},
		{Key: "up", Command: "cursor-up", Context: ContextNotesEditor},
		{Key: "down", Command: "cursor-down", Context: ContextNotesEditor},
		{Key: "left", Command: "cursor-left", Context: ContextNotesEditor},
		{Key: "right", Command: "cursor-right", Context: ContextNotesEditor},
		{Key: "ctrl+n", Command: "cursor-down", Context: ContextNotesEditor},
		{Key: "ctrl+p", Command: "cursor-up", Context: ContextNotesEditor},
		{Key: "home", Command: "line-start", Context: ContextNotesEditor},
		{Key: "end", Command: "line-end", Context: ContextNotesEditor},
		{Key: "ctrl+a", Command: "line-start", Context: ContextNotesEditor},
		{Key: "ctrl+e", Command: "line-end", Context: ContextNotesEditor},

		// Notes task modal context
		{Key: "enter", Command: "create-task", Context: ContextNotesTaskModal},
		{Key: "esc", Command: "cancel", Context: ContextNotesTaskModal},
		{Key: "tab", Command: "next-field", Context: ContextNotesTaskModal},
		{Key: "shift+tab", Command: "prev-field", Context: ContextNotesTaskModal},
	}
}

// Category represents a command category.
type Category string

const (
	CategoryNavigation Category = "Navigation"
	CategoryActions    Category = "Actions"
	CategoryView       Category = "View"
	CategorySearch     Category = "Search"
	CategorySystem     Category = "System"
)

// RegisterDefaults registers all default bindings with the given registry.
func RegisterDefaults(r *Registry) {
	for _, binding := range DefaultBindings() {
		r.RegisterBinding(binding)
	}
}
