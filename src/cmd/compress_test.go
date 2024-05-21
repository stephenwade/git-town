package cmd_test

import (
	"testing"

	"github.com/git-town/git-town/v14/src/cmd"
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	. "github.com/git-town/git-town/v14/src/gohacks/prelude"
	"github.com/shoenig/test/must"
)

func TestCompress(t *testing.T) {
	t.Parallel()
	t.Run("PreviousBranchAfterCompress", func(t *testing.T) {
		t.Parallel()
		t.Run("previous branch exists", func(t *testing.T) {
			t.Parallel()
			main := gitdomain.NewLocalBranchName("main")
			oldPrevious := gitdomain.NewLocalBranchName("previous")
			allBranches := gitdomain.BranchInfos{
				gitdomain.BranchInfo{
					LocalName:  main,
					SyncStatus: gitdomain.SyncStatusUpToDate,
				},
				gitdomain.BranchInfo{
					LocalName:  oldPrevious,
					SyncStatus: gitdomain.SyncStatusNotInSync,
				},
				gitdomain.BranchInfo{
					LocalName:  "other",
					SyncStatus: gitdomain.SyncStatusUpToDate,
				},
			}
			have := cmd.PreviousBranchAfterCompress(Some(oldPrevious), main, allBranches)
			want := Some(oldPrevious)
			must.Eq(t, want, have)
		})
	})
}
