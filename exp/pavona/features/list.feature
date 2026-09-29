Feature: List built-in templates
  Pavona can list all available built-in templates.

  Scenario: List shows all built-in templates
    When I run pavona with "list"
    * the output should contain "tool"
    * the output should contain "lib"
    * the output should contain "site"
    * the output should contain "tui"
    * the output should contain "app"
    * the output should contain "agent"
    * the output should contain "pavona"
    * the output should contain "monorepo-go"
