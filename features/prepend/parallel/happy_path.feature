Feature: Creating an additional parent branch

  Background:
    Given a feature branch "branch-1"
    And a feature branch "branch-2" as a child of "branch-1"
    And the current branch is "branch-2"
    And the commits
      | BRANCH   | LOCATION      | MESSAGE         |
      | branch-1 | local, origin | branch-1 commit |
      | branch-2 | local, origin | branch-2 commit |
    When I run "git-town prepend --parallel branch-3"

  Scenario: result
    Then it runs the commands
      | BRANCH | COMMAND                  |
      | old    | git fetch --prune --tags |
    And the current branch is now "parent-3"
    And these commits exist now
      | BRANCH   | LOCATION      | MESSAGE         |
      | branch-1 | local, origin | branch-1 commit |
      | branch-2 | local, origin | branch-2 commit |
      | branch-3 | local, origin | branch-3 commit |
    And this lineage exists now
      | BRANCH   | PARENT   |
      | branch-1 | main     |
      | branch-2 | branch-1 |
      | branch-3 | branch-1 |

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH | COMMAND              |
      | parent | git checkout old     |
      | old    | git branch -D parent |
    And the current branch is now "branch-2"
    And the initial commits exist
    And the initial lineage exists
