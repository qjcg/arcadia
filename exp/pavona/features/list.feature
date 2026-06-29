Feature: List built-in templates
  Pavona can list all available built-in templates.

  Scenario: List shows all built-in templates
    When I run pavona with "-l"
    Then the output should contain "tool"
    And the output should contain "lib"
    And the output should contain "site"
    And the output should contain "tui"
    And the output should contain "app"
    And the output should contain "agent"
