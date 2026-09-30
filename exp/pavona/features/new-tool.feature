Feature: Tool template
  The tool template generates a Go CLI project with cobra and BDD tests.

  Scenario: Create tool project
    When I create a project from the "tool" template with name "my-cli"
    Then the output directory should contain:
    | main.go       |
    | go.mod        |
    | Taskfile.yaml |
    | features      |
