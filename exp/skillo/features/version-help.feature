Feature: Version and help
  The CLI reports its version and usage instructions.

  Scenario: Version flag shows version string
    When I run "skillo" with "--version"
    Then the output should contain "skillo version 0.0.1"

  Scenario: Help flag shows available commands
    When I run "skillo" with "--help"
    Then the output should contain "Agent skills manager"
    And the output should contain "init"
    And the output should contain "get"
    And the output should contain "list"
    And the output should contain "search"
