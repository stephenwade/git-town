Feature: delete branches that were shipped or removed on another machine

  Background:
    Given the feature branches "active" and "old"
    And the commits
      | BRANCH | LOCATION      | MESSAGE       |
      | active | local, origin | active commit |
      | old    | local, origin | old commit    |
    And origin ships the "old" branch
    And the current branch is "old"
    And an uncommitted file
    When I run "git-town prune-branches"

  Scenario: result
    Then it runs the commands
      | BRANCH | COMMAND                  |
      | old    | git fetch --prune --tags |
      |        | git add -A               |
      |        | git stash                |
      |        | git checkout main        |
      | main   | git rebase origin/main   |
      |        | git branch -d old        |
      |        | git stash pop            |
    And the current branch is now "main"
    And the uncommitted file still exists
    And the branches are now
      | REPOSITORY    | BRANCHES     |
      | local, origin | main, active |
    And this branch lineage exists now
      | BRANCH | PARENT |
      | active | main   |

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH | COMMAND                                     |
      | main   | git add -A                                  |
      |        | git stash                                   |
      |        | git reset --hard {{ sha 'Initial commit' }} |
      |        | git branch old {{ sha 'old commit' }}       |
      |        | git checkout old                            |
      | old    | git stash pop                               |
    And the current branch is now "old"
    And the uncommitted file still exists
    And the initial branches and hierarchy exist
