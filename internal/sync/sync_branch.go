package sync

import (
	"github.com/git-town/git-town/v16/internal/config/configdomain"
	"github.com/git-town/git-town/v16/internal/git/gitdomain"
	"github.com/git-town/git-town/v16/internal/vm/opcodes"
	"github.com/git-town/git-town/v16/internal/vm/program"
	. "github.com/git-town/git-town/v16/pkg/prelude"
)

// BranchProgram syncs the given branch.
func BranchProgram(branch gitdomain.BranchInfo, args BranchProgramArgs) {
	parentOtherWorktree := false
	localName, hasLocalName := branch.LocalName.Get()
	parent, hasParent := args.Config.Lineage.Parent(localName).Get()
	if hasLocalName && hasParent {
		parentBranchInfo, hasParentBranchInfo := args.BranchInfos.FindByLocalName(parent).Get()
		parentOtherWorktree = hasParentBranchInfo && parentBranchInfo.SyncStatus == gitdomain.SyncStatusOtherWorktree
	}
	switch {
	case branch.SyncStatus == gitdomain.SyncStatusDeletedAtRemote:
		syncDeletedBranchProgram(args.Program, localName, parentOtherWorktree, args)
	case branch.SyncStatus == gitdomain.SyncStatusOtherWorktree:
		// Git Town doesn't sync branches that are active in another worktree
	default:
		ExistingBranchProgram(args.Program, branch, parent.BranchName(), parentOtherWorktree, args)
	}
	args.Program.Value.Add(&opcodes.EndOfBranchProgram{})
}

type BranchProgramArgs struct {
	BranchInfos        gitdomain.BranchInfos
	Config             configdomain.ValidatedConfig
	FirstCommitMessage Option[gitdomain.CommitMessage]
	InitialBranch      gitdomain.LocalBranchName
	Program            Mutable[program.Program]
	PushBranches       configdomain.PushBranches
	Remotes            gitdomain.Remotes
}

// ExistingBranchProgram provides the program to sync a particular branch.
func ExistingBranchProgram(list Mutable[program.Program], branch gitdomain.BranchInfo, parent gitdomain.BranchName, parentOtherWorktree bool, args BranchProgramArgs) {
	localName, hasLocalName := branch.LocalName.Get()
	if !hasLocalName {
		return
	}
	isMainOrPerennialBranch := args.Config.IsMainOrPerennialBranch(localName)
	if isMainOrPerennialBranch && !args.Remotes.HasOrigin() {
		// perennial branch but no remote --> this branch cannot be synced
		return
	}
	list.Value.Add(&opcodes.Checkout{Branch: localName})
	branchType := args.Config.BranchType(localName)
	switch branchType {
	case configdomain.BranchTypeFeatureBranch:
		FeatureBranchProgram(featureBranchArgs{
			firstCommitMessage:  args.FirstCommitMessage,
			localName:           localName,
			offline:             args.Config.Offline,
			parent:              parent,
			parentOtherWorktree: parentOtherWorktree,
			program:             list,
			pushBranches:        args.PushBranches,
			remoteName:          branch.RemoteName,
			syncStrategy:        args.Config.SyncFeatureStrategy.SyncStrategy(),
		})
	case configdomain.BranchTypePerennialBranch, configdomain.BranchTypeMainBranch:
		PerennialBranchProgram(branch, args)
	case configdomain.BranchTypeParkedBranch:
		ParkedBranchProgram(args.InitialBranch, featureBranchArgs{
			firstCommitMessage:  args.FirstCommitMessage,
			localName:           localName,
			offline:             args.Config.Offline,
			parent:              parent,
			parentOtherWorktree: parentOtherWorktree,
			program:             list,
			pushBranches:        args.PushBranches,
			remoteName:          branch.RemoteName,
			syncStrategy:        args.Config.SyncFeatureStrategy.SyncStrategy(),
		})
	case configdomain.BranchTypeContributionBranch:
		ContributionBranchProgram(args.Program, branch)
	case configdomain.BranchTypeObservedBranch:
		ObservedBranchProgram(branch.RemoteName, args.Program)
	case configdomain.BranchTypePrototypeBranch:
		FeatureBranchProgram(featureBranchArgs{
			firstCommitMessage:  args.FirstCommitMessage,
			localName:           localName,
			offline:             args.Config.Offline,
			parent:              parent,
			parentOtherWorktree: parentOtherWorktree,
			program:             list,
			pushBranches:        false,
			remoteName:          branch.RemoteName,
			syncStrategy:        args.Config.SyncPrototypeStrategy.SyncStrategy(),
		})
	}
	if args.PushBranches.IsTrue() && args.Remotes.HasOrigin() && args.Config.IsOnline() && branchType.ShouldPush(localName == args.InitialBranch) {
		switch {
		case !branch.HasTrackingBranch():
			list.Value.Add(&opcodes.CreateTrackingBranch{Branch: localName})
		case isMainOrPerennialBranch:
			list.Value.Add(&opcodes.PushCurrentBranch{CurrentBranch: localName})
		default:
			pushFeatureBranchProgram(list, localName, args.Config.SyncFeatureStrategy)
		}
	}
}

// pullParentBranchOfCurrentFeatureBranchOpcode adds the opcode to pull updates from the parent branch of the current feature branch into the current feature branch.
func pullParentBranchOfCurrentFeatureBranchOpcode(args pullParentBranchOfCurrentFeatureBranchOpcodeArgs) {
	switch args.syncStrategy {
	case configdomain.SyncFeatureStrategyMerge:
		args.program.Value.Add(&opcodes.MergeParent{CurrentBranch: args.branch, ParentActiveInOtherWorktree: args.parentOtherWorktree})
	case configdomain.SyncFeatureStrategyRebase:
		args.program.Value.Add(&opcodes.RebaseParent{CurrentBranch: args.branch, ParentActiveInOtherWorktree: args.parentOtherWorktree})
	case configdomain.SyncFeatureStrategyCompress:
		args.program.Value.Add(&opcodes.MergeParent{CurrentBranch: args.branch, ParentActiveInOtherWorktree: args.parentOtherWorktree})
	}
}

type pullParentBranchOfCurrentFeatureBranchOpcodeArgs struct {
	branch              gitdomain.LocalBranchName
	parentOtherWorktree bool
	program             Mutable[program.Program]
	syncStrategy        configdomain.SyncFeatureStrategy
}

func pushFeatureBranchProgram(list Mutable[program.Program], branch gitdomain.LocalBranchName, syncFeatureStrategy configdomain.SyncFeatureStrategy) {
	switch syncFeatureStrategy {
	case configdomain.SyncFeatureStrategyMerge:
		list.Value.Add(&opcodes.PushCurrentBranch{CurrentBranch: branch})
	case configdomain.SyncFeatureStrategyRebase:
		list.Value.Add(&opcodes.ForcePushCurrentBranch{ForceIfIncludes: true})
	case configdomain.SyncFeatureStrategyCompress:
		list.Value.Add(&opcodes.ForcePushCurrentBranch{ForceIfIncludes: false})
	}
}

// updateCurrentPerennialBranchOpcode provides the opcode to update the current perennial branch with changes from the given other branch.
func updateCurrentPerennialBranchOpcode(list Mutable[program.Program], otherBranch gitdomain.RemoteBranchName, strategy configdomain.SyncPerennialStrategy) {
	switch strategy {
	case configdomain.SyncPerennialStrategyMerge:
		list.Value.Add(&opcodes.Merge{Branch: otherBranch.BranchName()})
	case configdomain.SyncPerennialStrategyRebase:
		list.Value.Add(&opcodes.RebaseBranch{Branch: otherBranch.BranchName()})
	}
}
