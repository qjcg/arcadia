Feature: Agent template
  The agent template generates a NATS Agent Protocol service.

  Scenario: Hydrate agent template
    When I hydrate the "agent" template with name "triagebot"
    Then the output directory should contain "main.go"
    And the output directory should contain "go.mod"
