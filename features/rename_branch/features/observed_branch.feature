Feature: rename an observed branch

  Background:
    Given a Git repo with origin
    And the branches
      | NAME     | TYPE     | PARENT | LOCATIONS     |
      | observed | observed | main   | local, origin |
    And the current branch is "observed"
    And the commits
      | BRANCH   | LOCATION      | MESSAGE               |
      | observed | local, origin | somebody elses commit |
    When I run "git-town rename-branch observed new"

  Scenario: result
    Then it runs the commands
      | BRANCH   | COMMAND                   |
      | observed | git fetch --prune --tags  |
      |          | git branch new observed   |
      |          | git checkout new          |
      | new      | git push -u origin new    |
      |          | git push origin :observed |
      |          | git branch -D observed    |
    And the current branch is now "new"
    And the observed branches are now "new"
    And these commits exist now
      | BRANCH | LOCATION      | MESSAGE               |
      | new    | local, origin | somebody elses commit |
    And this lineage exists now
      | BRANCH | PARENT |

  Scenario: undo
    When I run "git-town undo"
    Then it runs the commands
      | BRANCH   | COMMAND                                               |
      | new      | git branch observed {{ sha 'somebody elses commit' }} |
      |          | git push -u origin observed                           |
      |          | git push origin :new                                  |
      |          | git checkout observed                                 |
      | observed | git branch -D new                                     |
    And the current branch is now "observed"
    And the observed branches are now "observed"
    And the initial commits exist
    And the initial branches and lineage exist
