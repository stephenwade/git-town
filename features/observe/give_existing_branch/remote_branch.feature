Feature: observe a remote branch

  Background:
    Given a Git repo with origin
    And the branch
      | NAME           | TYPE    | PARENT | LOCATIONS |
      | remote-feature | feature | main   | origin    |
    And I run "git fetch"
    And an uncommitted file
    When I run "git-town observe remote-feature"

  Scenario: result
    Then it runs the commands
      | BRANCH | COMMAND                     |
      |        | git checkout remote-feature |
    And it prints:
      """
      branch "remote-feature" is now an observed branch
      """
    And the current branch is now "remote-feature"
    And branch "remote-feature" is now observed
    And the uncommitted file still exists

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH         | COMMAND                      |
      | remote-feature | git add -A                   |
      |                | git stash                    |
      |                | git checkout main            |
      | main           | git branch -D remote-feature |
      |                | git stash pop                |
    And the current branch is now "main"
    And there are now no observed branches
    And the uncommitted file still exists
