Feature: Version flag
  Pavona reports its version.

  Scenario: Version flag prints version
    When I run pavona with "--version"
    Then the output should contain "pavona"
