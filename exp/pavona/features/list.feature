Feature: List built-in templates
  Pavona can list all available built-in templates.

  Scenario: List shows all built-in templates
    When I run pavona with "list"
    Then the output should contain:
    | tool        |
    | lib         |
    | site        |
    | tui         |
    | app         |
    | agent       |
    | pavona      |
    | monorepo-go |
