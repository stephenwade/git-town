# git town append

> _git town append [--prototype] &lt;branch-name&gt;_

The _append_ command creates a new feature branch with the given name as a
direct child of the current branch and brings over all uncommitted changes to
the new branch.

When running without uncommitted changes in your workspace, it also
[syncs](sync.md) the current branch to ensure your work in the new branch
happens on top of the current state of the repository. If the workspace contains
uncommitted changes, `git append` does not perform this sync to let you commit
your open changes first and then sync manually.

### Positional argument

When given a non-existing branch name, `git append` creates a new feature branch
with the main branch as its parent.

Consider this branch setup:

```
main
 \
* feature-1
```

We are on the `feature-1` branch. After running `git append feature-2`, our
repository will have this branch setup:

```
main
 \
  feature-1
   \
*   feature-2
```

### --prototype / -p

Adding the `--prototype` aka `-p` switch creates a
[prototype branch](../branch-types.md#prototype-branches)).

### Configuration

If [push-new-branches](../preferences/push-new-branches.md) is set, `git append`
also creates the tracking branch for the new feature branch. This behavior is
disabled by default to make `git append` run fast and save CI runs. The first
run of `git sync` will create the remote tracking branch.

If the configuration setting
[create-prototype-branches](../preferences/create-prototype-branches.md) is set,
`git append` always creates a
[prototype branch](../branch-types.md#prototype-branches).
