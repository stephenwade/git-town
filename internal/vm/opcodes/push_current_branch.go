package opcodes

import (
	"github.com/git-town/git-town/v16/internal/git/gitdomain"
	"github.com/git-town/git-town/v16/internal/vm/shared"
)

// PushCurrentBranch pushes the current branch to its existing tracking branch.
type PushCurrentBranch struct {
	CurrentBranch           gitdomain.LocalBranchName
	undeclaredOpcodeMethods `exhaustruct:"optional"`
}

func (self *PushCurrentBranch) Run(args shared.RunArgs) error {
	shouldPush, err := args.Git.ShouldPushBranch(args.Backend, self.CurrentBranch)
	if err != nil {
		return err
	}
	if !shouldPush {
		return nil
	}
	return args.Git.PushCurrentBranch(args.Frontend, args.Config.Config.NoPushHook())
}
