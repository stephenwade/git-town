package sync

import (
	"github.com/git-town/git-town/v14/src/cmd/cmdhelpers"
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	. "github.com/git-town/git-town/v14/src/gohacks/prelude"
	"github.com/git-town/git-town/v14/src/vm/opcodes"
)

// BranchesProgram syncs all given branches.
func BranchesProgram(args BranchesProgramArgs) {
	for _, branch := range args.BranchesToSync {
		BranchProgram(branch, args.BranchProgramArgs)
	}
	finalBranchCandidates := gitdomain.LocalBranchNames{args.InitialBranch}
	if previousBranch, hasPreviousBranch := args.PreviousBranch.Get(); hasPreviousBranch {
		finalBranchCandidates = append(finalBranchCandidates, previousBranch)
	}
	args.Program.Add(&opcodes.CheckoutFirstExisting{
		Branches:   finalBranchCandidates,
		MainBranch: args.Config.MainBranch,
	})
	if args.Remotes.HasOrigin() && args.ShouldPushTags && args.Config.IsOnline() {
		args.Program.Add(&opcodes.PushTags{})
	}
	cmdhelpers.Wrap(args.Program, cmdhelpers.WrapOptions{
		DryRun:           args.DryRun,
		RunInGitRoot:     true,
		StashOpenChanges: args.HasOpenChanges,
		PreviousBranch:   previousBranchAfterSync(args.PreviousBranch, args.Config.MainBranch, args.BranchInfos),
	})
}

type BranchesProgramArgs struct {
	BranchProgramArgs
	BranchesToSync gitdomain.BranchInfos
	DryRun         bool
	HasOpenChanges bool
	InitialBranch  gitdomain.LocalBranchName
	PreviousBranch Option[gitdomain.LocalBranchName]
	ShouldPushTags bool
}

func previousBranchAfterSync(oldPreviousBranch Option[gitdomain.LocalBranchName], mainBranch gitdomain.LocalBranchName, allBranches gitdomain.BranchInfos) Option[gitdomain.LocalBranchName] {
	mainInfo := allBranches.FindByLocalName(mainBranch).GetOrPanic()
	var mainBranchOpt Option[gitdomain.LocalBranchName]
	if mainInfo.SyncStatus != gitdomain.SyncStatusOtherWorktree {
		mainBranchOpt = Some(mainBranch)
	} else {
		mainBranchOpt = None[gitdomain.LocalBranchName]()
	}
	oldPrevious, hasOldPrevious := oldPreviousBranch.Get()
	if !hasOldPrevious {
		return mainBranchOpt
	}
	oldPreviousInfo, hasOldPreviousInfo := allBranches.FindByLocalName(oldPrevious).Get()
	if !hasOldPreviousInfo {
		return mainBranchOpt
	}
	if oldPreviousInfo.SyncStatus == gitdomain.SyncStatusOtherWorktree {
		return mainBranchOpt
	}
	return Some(oldPrevious)
}
