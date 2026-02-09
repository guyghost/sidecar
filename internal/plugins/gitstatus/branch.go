package gitstatus

import "github.com/guyghost/sidecar/internal/git"

// Re-export branch types from internal/git for backward compatibility.
type (
	Branch      = git.Branch
	BranchError = git.BranchError
)

// Re-export branch functions.
var (
	GetBranches       = git.GetBranches
	CheckoutBranch    = git.CheckoutBranch
	CreateBranch      = git.CreateBranch
	DeleteBranch      = git.DeleteBranch
	ForceDeleteBranch = git.ForceDeleteBranch
)
