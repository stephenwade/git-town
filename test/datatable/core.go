package datatable

import "github.com/git-town/git-town/v16/internal/git/gitdomain"

type runner interface {
	SHAsForCommit(name string) gitdomain.SHAs
}
