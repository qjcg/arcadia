Feature: Custom template
  Pavona can create a project from a custom template in a local directory.

  Scenario: Create project from custom template
    Given a custom template with config.cue and main.go.tmpl
    When I create a project from the custom template with name "my-custom"
    Then the output directory should contain "main.go"
    And "main.go" should contain "Hello, World!"
