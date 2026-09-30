Feature: Pavona template
  The pavona template creates a starter template for generating projects.

  Scenario: Create pavona template
    When I create a project from the "pavona" template with name "my-template"
    Then the output directory should contain "config.cue"
    And "config.cue" should contain:
    | my-template          |
    | project_name: string |
    | greeting?: string    |
