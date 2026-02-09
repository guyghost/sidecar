package gitstatus

import "github.com/marcus/sidecar/internal/git"

// Re-export remote types from internal/git for backward compatibility.
type RemoteError = git.RemoteError

// Re-export remote functions.
var (
	ExecuteFetch         = git.ExecuteFetch
	ExecutePull          = git.ExecutePull
	ExecutePullRebase    = git.ExecutePullRebase
	ExecutePullFFOnly    = git.ExecutePullFFOnly
	ExecutePullAutostash = git.ExecutePullAutostash
	GetConflictedFiles   = git.GetConflictedFiles
	IsConflictError      = git.IsConflictError
	AbortMerge           = git.AbortMerge
	AbortRebase          = git.AbortRebase
	IsRebaseInProgress   = git.IsRebaseInProgress
)
