package opcodes

import (
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	. "github.com/git-town/git-town/v14/src/gohacks/prelude"
	"github.com/git-town/git-town/v14/src/vm/shared"
)

// PreserveCheckoutHistory does stuff.
type PreserveCheckoutHistory struct {
	PreviousBranch          Option[gitdomain.LocalBranchName]
	undeclaredOpcodeMethods `exhaustruct:"optional"`
}

func (self *PreserveCheckoutHistory) Run(args shared.RunArgs) error {
	if !args.Backend.CurrentBranchCache.Initialized() {
		// the branch cache is not initialized --> there were no branch changes --> no need to restore the branch history
		return nil
	}
	currentBranch := args.Backend.CurrentBranchCache.Value()
	actualPreviousBranch := args.Backend.CurrentBranchCache.Previous()
	wantPreviousBranch, hasWantPrevious := self.PreviousBranch.Get()
	if !hasWantPrevious {
		return nil
	}
	if actualPreviousBranch == wantPreviousBranch {
		return nil
	}
	if err := args.Backend.CheckoutBranchUncached(wantPreviousBranch); err != nil {
		return err
	}
	return args.Backend.CheckoutBranchUncached(currentBranch)
}
