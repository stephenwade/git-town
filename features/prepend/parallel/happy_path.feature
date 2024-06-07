Feature: Creating an additional parent branch

  Background:
    Given a feature branch "old-parent"
    And a feature branch "child" as a child of "old-parent"
    And the current branch is "child"
    And the commits
      | BRANCH     | LOCATION      | MESSAGE           |
      | old-parent | local, origin | old-parent commit |
      | child      | local, origin | child commit      |
    When I run "git-town prepend --parallel new-parent"

  @debug @this
  Scenario: result
    Then it runs the commands
      | BRANCH     | COMMAND                                    |
      | child      | git fetch --prune --tags                   |
      |            | git checkout main                          |
      | main       | git rebase origin/main                     |
      |            | git checkout old-parent                    |
      | old-parent | git merge --no-edit --ff origin/old-parent |
      |            | git merge --no-edit --ff main              |
      |            | git checkout child                         |
      | child      | git merge --no-edit --ff origin/child      |
      |            | git merge --no-edit --ff old-parent        |
      |            | git push                                   |
      |            | git checkout -b new-parent main            |
    And the current branch is now "new-parent"
    And these commits exist now
      | BRANCH     | LOCATION      | MESSAGE                              |
      | child      | local, origin | child commit                         |
      |            |               | old-parent commit                    |
      |            |               | Merge branch 'old-parent' into child |
      | old-parent | local, origin | old-parent commit                    |
    And this lineage exists now
      | BRANCH     | PARENT                 |
      | child      | old-parent, new-parent |
      | new-parent | main                   |
      | old-parent | main                   |

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH | COMMAND              |
      | parent | git checkout old     |
      | old    | git branch -D parent |
    And the current branch is now "branch-2"
    And the initial commits exist
    And the initial lineage exists
