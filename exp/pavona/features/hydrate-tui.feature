Feature: TUI template
  The TUI template generates a terminal UI project.

  Scenario: Hydrate tui template
    When I hydrate the "tui" template with name "chatmonitor"
    Then the output directory should contain "main.go"
    And the output directory should contain "go.mod"
