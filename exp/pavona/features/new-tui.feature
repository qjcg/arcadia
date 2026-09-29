Feature: TUI template
  The TUI template generates a terminal UI project.

  Scenario: Create TUI project
    When I create a project from the "tui" template with name "chatmonitor"
    Then the output directory should contain "main.go"
    And the output directory should contain "go.mod"
