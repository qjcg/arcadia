Feature: Pavona template
  The pavona template creates a starter template that can be hydrated in turn.

  Scenario: Hydrate pavona template
    When I hydrate the "pavona" template with name "my-template"
    Then the output directory should contain "config.cue"
    And "config.cue" should contain "my-template"
    And "config.cue" should contain "project_name: string"
    And "config.cue" should contain "greeting?: string"
