package configdomain

import (
	"github.com/git-town/git-town/v14/src/git/gitdomain"
	. "github.com/git-town/git-town/v14/src/gohacks/prelude"
	"github.com/git-town/git-town/v14/src/gohacks/slice"
)

// Lineage encapsulates all data and functionality around parent branches.
// branch --> its parent
// Lineage only contains branches that have ancestors.
type Lineage struct {
	data []LineageEntry
}

func NewLineage() Lineage {
	return Lineage{
		data: make([]LineageEntry, 0),
	}
}

func (self *Lineage) Add(branch, parent gitdomain.LocalBranchName) {
	if self.HasParent(branch, parent) {
		return
	}
	self.data = append(self.data, LineageEntry{
		Child:  branch,
		Parent: parent,
	})
}

// Ancestors provides the names of all parent branches of the branch with the given name.
func (self Lineage) Ancestors(branch gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	result := gitdomain.LocalBranchNames{}
	self.addAncestors(branch, &result)
	return result
}

// adds all ancestors of the given branch to the given ancestors list
func (self Lineage) addAncestors(branch gitdomain.LocalBranchName, ancestors *gitdomain.LocalBranchNames) {
	for _, parent := range self.Parents(branch) {
		*ancestors = append(*ancestors, parent)
		self.addAncestors(parent, ancestors)
	}
}

// AncestorsWithoutRoot provides the names of all parent branches of the branch with the given name, excluding the root perennial branch.
func (self Lineage) AncestorsWithoutRoot(branch gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	ancestors := self.Ancestors(branch)
	if len(ancestors) > 0 {
		return ancestors[1:]
	}
	return ancestors
}

// BranchAndAncestors provides the full ancestry for the branch with the given name,
// including the branch.
func (self Lineage) BranchAndAncestors(branchName gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	return append(self.Ancestors(branchName), branchName)
}

// BranchLineageWithoutRoot provides all branches in the lineage of the given branch,
// from oldest to youngest, including the given branch.
func (self Lineage) BranchLineageWithoutRoot(branch gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	if !self.HasParents(branch) {
		return self.Descendants(branch)
	}
	return append(append(self.AncestorsWithoutRoot(branch), branch), self.Descendants(branch)...)
}

// BranchNames provides the names of all branches in this Lineage, sorted alphabetically.
func (self Lineage) BranchNames() gitdomain.LocalBranchNames {
	result := gitdomain.LocalBranchNames(self.Branches())
	result.Sort()
	return result
}

// provides all branches for which the parent is known
func (self Lineage) Branches() gitdomain.LocalBranchNames {
	result := make(gitdomain.LocalBranchNames, 0, self.Len())
	for e, entry := range self.Entries() {
		result[e] = entry.Child
	}
	return result
}

// BranchesAndAncestors provides the full lineage for the branches with the given names,
// including the branches themselves.
func (self Lineage) BranchesAndAncestors(branchNames gitdomain.LocalBranchNames) gitdomain.LocalBranchNames {
	result := branchNames
	for _, branchName := range branchNames {
		result = slice.AppendAllMissing(result, self.Ancestors(branchName)...)
	}
	return self.OrderHierarchically(result)
}

// Children provides the names of all branches that have the given branch as their parent.
func (self Lineage) Children(branch gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	result := gitdomain.LocalBranchNames{}
	for _, entry := range self.Entries() {
		if entry.Parent == branch {
			result = append(result, entry.Child)
		}
	}
	result.Sort()
	return result
}

// Descendants provides all branches that depend on the given branch in its lineage.
func (self Lineage) Descendants(branch gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	result := gitdomain.LocalBranchNames{}
	for _, child := range self.Children(branch) {
		result = append(result, child)
		result = append(result, self.Descendants(child)...)
	}
	return result
}

func (self Lineage) Entries() []LineageEntry {
	return self.data
}

// indicates whether the given branch has the given parent
func (self Lineage) HasParent(branch, parent gitdomain.LocalBranchName) bool {
	for _, entry := range self.data {
		if entry.Child == branch && entry.Parent == parent {
			return true
		}
	}
	return false
}

// HasParents returns whether or not the given branch has at least one parent.
func (self Lineage) HasParents(branch gitdomain.LocalBranchName) bool {
	for _, entry := range self.Entries() {
		if entry.Child == branch {
			return true
		}
	}
	return false
}

// IsAncestor indicates whether the given branch is an ancestor of the other given branch.
func (self Lineage) IsAncestor(ancestor, other gitdomain.LocalBranchName) bool {
	ancestors := self.Ancestors(other)
	return ancestors.Contains(ancestor)
}

func (self Lineage) IsEmpty() bool {
	return self.data == nil || self.Len() == 0
}

func (self Lineage) Len() int {
	return len(self.data)
}

// OrderHierarchically provides the given branches sorted so that ancestor branches come before their descendants.
func (self Lineage) OrderHierarchically(branches gitdomain.LocalBranchNames) gitdomain.LocalBranchNames {
	result := make(gitdomain.LocalBranchNames, 0, len(self.data))
	for _, root := range self.Roots() {
		self.addChildrenHierarchically(&result, root, branches)
	}
	result = result.AppendAllMissing(branches...)
	return result
}

// Parent provides the name of the parent branch for the given branch or nil if the branch has no parent.
func (self Lineage) Parents(branch gitdomain.LocalBranchName) gitdomain.LocalBranchNames {
	result := gitdomain.LocalBranchNames{}
	for _, entry := range self.Entries() {
		if entry.Child == branch {
			result = append(result, entry.Parent)
		}
	}
	return result
}

// provides the LineageEntry instances that have the given branch as the parent
func (self Lineage) ChildEntries(branch gitdomain.LocalBranchName) []*LineageEntry {
	result := []*LineageEntry{}
	for e, entry := range self.data {
		if entry.Parent == branch {
			result = append(result, &self.data[e])
		}
	}
	return result
}

func (self *Lineage) removeMatching(query lineageQuery) {
	newData := make([]LineageEntry, 0, self.Len())
	queryChild, hasQueryChild := query.child.Get()
	queryParent, hasQueryParent := query.parent.Get()
	for _, entry := range self.data {
		entryMatches := false
		if hasQueryChild && entry.Child == queryChild {
			entryMatches = true
		}
		if hasQueryParent && entry.Parent == queryParent {
			entryMatches = true
		}
		if !entryMatches {
			newData = append(newData, entry)
		}
	}
	self.data = newData
}

type lineageQuery struct {
	child  Option[gitdomain.LocalBranchName]
	parent Option[gitdomain.LocalBranchName]
}

// RemoveBranch removes the given branch completely from this lineage.
func (self Lineage) RemoveBranch(branch gitdomain.LocalBranchName) {
	parents := self.Parents(branch)
	children := self.Children(branch)
	self.removeMatching(lineageQuery{
		child:  Some(branch),
		parent: Some(branch),
	})
	for _, child := range children {
		for _, parent := range parents {
			self.Add(child, parent)
		}
	}
}

// Roots provides the branches with children and no parents.
func (self Lineage) Roots() gitdomain.LocalBranchNames {
	roots := gitdomain.LocalBranchNames{}
	for _, entry := range self.data {
		hasParent := self.HasParents(entry.Parent)
		if !hasParent && !slice.Contains(roots, entry.Child) {
			roots = append(roots, entry.Child)
		}
	}
	roots.Sort()
	return roots
}

func (self Lineage) addChildrenHierarchically(result *gitdomain.LocalBranchNames, currentBranch gitdomain.LocalBranchName, allBranches gitdomain.LocalBranchNames) {
	if allBranches.Contains(currentBranch) {
		*result = append(*result, currentBranch)
	}
	for _, child := range self.Children(currentBranch) {
		self.addChildrenHierarchically(result, child, allBranches)
	}
}
