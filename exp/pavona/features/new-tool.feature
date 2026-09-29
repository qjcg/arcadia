Feature: Tool template
  The tool template generates a Go CLI project with cobra and BDD tests.

  Scenario: Create tool project
    When I create a project from the "tool" template with name "my-cli"
    Then the output directory should contain "main.go"
    And the output directory should contain "go.mod"
    And the output directory should contain "Taskfile.yaml"
    And the output directory should contain "features"
