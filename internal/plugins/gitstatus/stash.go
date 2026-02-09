package gitstatus

import "github.com/marcus/sidecar/internal/git"

// Re-export stash types from internal/git for backward compatibility.
type (
	Stash      = git.Stash
	StashList  = git.StashList
	StashError = git.StashError
)

// Re-export stash functions.
var (
	GetStashList              = git.GetStashList
	StashPush                 = git.StashPush
	StashPushWithMessage      = git.StashPushWithMessage
	StashPushIncludeUntracked = git.StashPushIncludeUntracked
	StashPop                  = git.StashPop
	StashPopRef               = git.StashPopRef
	StashApply                = git.StashApply
	StashDrop                 = git.StashDrop
)
