package keymap

import (
	"errors"
	"strings"
)

// FocusContext represents a named scope for key bindings.
// Contexts enable the same key to have different meanings depending on
// where the user is in the application (e.g., "global" vs "git-status").
type FocusContext string

// FocusContext constants - compile-time discoverable, typed contexts.
// These are used as values for the Context field in Binding structs
// and returned by the Plugin.FocusContext() method.
const (
	// Global context - always active as fallback
	ContextGlobal FocusContext = "global"

	// Project switcher context
	ContextProjectSwitcher FocusContext = "project-switcher"

	// Git status contexts
	ContextGitNoRepo        FocusContext = "git-no-repo"
	ContextGitStatus        FocusContext = "git-status"
	ContextGitStatusCommits FocusContext = "git-status-commits"
	ContextGitStatusDiff    FocusContext = "git-status-diff"

	// Git modal contexts
	ContextGitHistorySearch FocusContext = "git-history-search"
	ContextGitPathFilter    FocusContext = "git-path-filter"
	ContextGitCommitPreview FocusContext = "git-commit-preview"
	ContextGitDiff          FocusContext = "git-diff"
	ContextGitPushMenu      FocusContext = "git-push-menu"
	ContextGitPullMenu      FocusContext = "git-pull-menu"
	ContextGitPullConflict  FocusContext = "git-pull-conflict"
	ContextGitError         FocusContext = "git-error"
	ContextGitStashPop      FocusContext = "git-stash-pop"
	ContextGitCommit        FocusContext = "git-commit"
	ContextGitHistory       FocusContext = "git-history"
	ContextGitCommitDetail  FocusContext = "git-commit-detail"

	// Issue contexts
	ContextIssueInput   FocusContext = "issue-input"
	ContextIssuePreview FocusContext = "issue-preview"

	// Conversations contexts
	ContextConversationsSidebar       FocusContext = "conversations-sidebar"
	ContextConversationsMain          FocusContext = "conversations-main"
	ContextConversationsSearch        FocusContext = "conversations-search"
	ContextConversationsFilter        FocusContext = "conversations-filter"
	ContextConversationsContentSearch FocusContext = "conversations-content-search"
	ContextConversationsResumeModal   FocusContext = "conversations-resume-modal"
	ContextTurnDetail                 FocusContext = "turn-detail"

	// File browser contexts
	ContextFileBrowserTree          FocusContext = "file-browser-tree"
	ContextFileBrowserPreview       FocusContext = "file-browser-preview"
	ContextFileBrowserSearch        FocusContext = "file-browser-search"
	ContextFileBrowserQuickOpen     FocusContext = "file-browser-quick-open"
	ContextFileBrowserProjectSearch FocusContext = "file-browser-project-search"
	ContextFileBrowserFileOp        FocusContext = "file-browser-file-op"
	ContextFileBrowserLineJump      FocusContext = "file-browser-line-jump"
	ContextFileBrowserContentSearch FocusContext = "file-browser-content-search"
	ContextFileBrowserInlineEdit    FocusContext = "file-browser-inline-edit"
	ContextFileBrowserInfo          FocusContext = "file-browser-info"
	ContextFileBrowserBlame         FocusContext = "file-browser-blame"

	// Workspace contexts
	ContextWorkspaceList               FocusContext = "workspace-list"
	ContextWorkspacePreview            FocusContext = "workspace-preview"
	ContextWorkspaceInteractive        FocusContext = "workspace-interactive"
	ContextWorkspaceCreate             FocusContext = "workspace-create"
	ContextWorkspaceTaskLink           FocusContext = "workspace-task-link"
	ContextWorkspaceMerge              FocusContext = "workspace-merge"
	ContextWorkspaceMergeError         FocusContext = "workspace-merge-error"
	ContextWorkspaceAgentChoice        FocusContext = "workspace-agent-choice"
	ContextWorkspaceConfirmDelete      FocusContext = "workspace-confirm-delete"
	ContextWorkspaceConfirmDeleteShell FocusContext = "workspace-confirm-delete-shell"
	ContextWorkspaceCommitForMerge     FocusContext = "workspace-commit-for-merge"
	ContextWorkspacePromptPicker       FocusContext = "workspace-prompt-picker"
	ContextWorkspaceRenameShell        FocusContext = "workspace-rename-shell"
	ContextWorkspaceTypeSelector       FocusContext = "workspace-type-selector"
	ContextWorkspaceFetchPR            FocusContext = "workspace-fetch-pr"
	ContextWorkspaceFilePicker         FocusContext = "workspace-file-picker"

	// Notes contexts
	ContextNotesList        FocusContext = "notes-list"
	ContextNotesInfo        FocusContext = "notes-info"
	ContextNotesSearch      FocusContext = "notes-search"
	ContextNotesPreview     FocusContext = "notes-preview"
	ContextNotesEditor      FocusContext = "notes-editor"
	ContextNotesTaskModal   FocusContext = "notes-task-modal"
	ContextNotesDeleteModal FocusContext = "notes-delete-modal"
	ContextNotesInlineEdit  FocusContext = "notes-inline-edit"

	// TD monitor contexts
	ContextTDMonitor      FocusContext = "td-monitor"
	ContextTDSearch       FocusContext = "td-search"
	ContextTDForm         FocusContext = "td-form"
	ContextTDBoardEditor  FocusContext = "td-board-editor"
	ContextTDConfirm      FocusContext = "td-confirm"
	ContextTDCloseConfirm FocusContext = "td-close-confirm"
	ContextTDBoard        FocusContext = "td-board"

	// Theme switcher context
	ContextThemeSwitcher FocusContext = "theme-switcher"
)

// AllContexts returns a slice of all defined focus contexts.
// Useful for validation and discovery.
func AllContexts() []FocusContext {
	return []FocusContext{
		ContextGlobal,
		ContextProjectSwitcher,
		ContextGitNoRepo,
		ContextGitStatus,
		ContextGitStatusCommits,
		ContextGitStatusDiff,
		ContextGitHistorySearch,
		ContextGitPathFilter,
		ContextGitCommitPreview,
		ContextGitDiff,
		ContextGitPushMenu,
		ContextGitPullMenu,
		ContextGitPullConflict,
		ContextGitError,
		ContextGitStashPop,
		ContextGitCommit,
		ContextGitHistory,
		ContextGitCommitDetail,
		ContextIssueInput,
		ContextIssuePreview,
		ContextConversationsSidebar,
		ContextConversationsMain,
		ContextConversationsSearch,
		ContextConversationsFilter,
		ContextConversationsContentSearch,
		ContextConversationsResumeModal,
		ContextTurnDetail,
		ContextFileBrowserTree,
		ContextFileBrowserPreview,
		ContextFileBrowserSearch,
		ContextFileBrowserQuickOpen,
		ContextFileBrowserProjectSearch,
		ContextFileBrowserFileOp,
		ContextFileBrowserLineJump,
		ContextFileBrowserContentSearch,
		ContextFileBrowserInlineEdit,
		ContextFileBrowserInfo,
		ContextFileBrowserBlame,
		ContextWorkspaceList,
		ContextWorkspacePreview,
		ContextWorkspaceInteractive,
		ContextWorkspaceCreate,
		ContextWorkspaceTaskLink,
		ContextWorkspaceMerge,
		ContextWorkspaceMergeError,
		ContextWorkspaceAgentChoice,
		ContextWorkspaceConfirmDelete,
		ContextWorkspaceConfirmDeleteShell,
		ContextWorkspaceCommitForMerge,
		ContextWorkspacePromptPicker,
		ContextWorkspaceRenameShell,
		ContextWorkspaceTypeSelector,
		ContextWorkspaceFetchPR,
		ContextWorkspaceFilePicker,
		ContextNotesList,
		ContextNotesInfo,
		ContextNotesSearch,
		ContextNotesPreview,
		ContextNotesEditor,
		ContextNotesTaskModal,
		ContextNotesDeleteModal,
		ContextNotesInlineEdit,
		ContextTDMonitor,
		ContextTDSearch,
		ContextTDForm,
		ContextTDBoardEditor,
		ContextTDConfirm,
		ContextTDCloseConfirm,
		ContextTDBoard,
		ContextThemeSwitcher,
	}
}

// String returns the string representation of the focus context.
// This implements the Stringer interface for easy debugging/logging.
func (fc FocusContext) String() string {
	return string(fc)
}

// IsRoot returns true if this context is a root view where 'q' should quit.
// Root contexts are plugin top-level views (not sub-views like detail/diff/commit).
func (fc FocusContext) IsRoot() bool {
	switch fc {
	case ContextGlobal, "":
		return true
	// Plugin root contexts where 'q' is not used for navigation
	case ContextConversationsSidebar, ContextConversationsMain, "conversations":
		return true
	case ContextGitStatus, ContextGitStatusCommits, ContextGitStatusDiff, ContextGitCommitPreview:
		return true
	case ContextFileBrowserTree, ContextFileBrowserPreview:
		return true
	case ContextWorkspaceList, ContextWorkspacePreview:
		return true
	case ContextTDMonitor, ContextTDBoard:
		return true
	case ContextNotesList:
		return true
	default:
		return false
	}
}

// IsTextInput returns true if this context is a text input mode
// where alphanumeric keys should be forwarded to the plugin for typing.
func (fc FocusContext) IsTextInput() bool {
	switch fc {
	case ContextTDSearch, ContextTDForm, ContextTDBoardEditor, ContextTDConfirm, ContextTDCloseConfirm,
		ContextThemeSwitcher,
		ContextIssueInput:
		return true
	default:
		return false
	}
}

// ParseContext converts a string to a FocusContext.
// Returns an error if the string does not match any known context.
// This is used for parsing user config keymap overrides which use string context names.
func ParseContext(s string) (FocusContext, error) {
	if s == "" {
		return "", errors.New("context string is empty")
	}

	// Try to match against all known contexts
	fc := FocusContext(s)
	for _, known := range AllContexts() {
		if fc == known {
			return fc, nil
		}
	}

	// Check if it's a valid context but not one of our constants
	// (e.g., dynamically registered contexts from plugins)
	if strings.TrimSpace(s) != "" {
		return fc, nil
	}

	return "", errors.New("invalid context: " + s)
}
