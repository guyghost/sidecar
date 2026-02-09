package gitstatus

import "github.com/guyghost/sidecar/internal/git"

// Re-export push types from internal/git for backward compatibility.
type (
	PushStatus = git.PushStatus
	PushError  = git.PushError
)

// Re-export push functions.
var (
	GetPushStatus          = git.GetPushStatus
	ExecutePush            = git.ExecutePush
	ExecutePushForce       = git.ExecutePushForce
	ExecutePushSetUpstream = git.ExecutePushSetUpstream
	GetRemoteName          = git.GetRemoteName
	HasRemote              = git.HasRemote
	ParsePushOutput        = git.ParsePushOutput
	IsPushRejectedError    = git.IsPushRejectedError
)

// isPushRejectedError wraps git.IsPushRejectedError for internal use.
func isPushRejectedError(err error) bool {
	return git.IsPushRejectedError(err)
}
