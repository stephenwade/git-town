Feature: already existing remote branch

  Background:
    Given a Git repo with origin
    And the branches
      | NAME     | TYPE    | PARENT | LOCATIONS     |
      | old      | feature | main   | local, origin |
      | existing | feature | main   | origin        |
    And the current branch is "old"
    When I run "git-town prepend existing"

  Scenario: result
    Then it runs the commands
      | BRANCH | COMMAND                  |
      | old    | git fetch --prune --tags |
    And it prints the error:
      """
      there is already a branch "existing" at the "origin" remote
      """

  Scenario: undo
    When I run "git-town undo"
    Then it runs no commands
    And the current branch is now "old"
    And the initial commits exist
    And the initial lineage exists
