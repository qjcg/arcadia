Feature: Custom template
  Pavona can hydrate a custom template from a local directory.

  Scenario: Hydrate custom template
    Given a custom template with config.cue and main.go.tmpl
    When I hydrate the custom template with name "my-custom"
    Then the output directory should contain "main.go"
    And "main.go" should contain "Hello, World!"
