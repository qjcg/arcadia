Feature: Agent template
  The agent template generates a NATS Agent Protocol service.

  Scenario: Create agent project
    When I create a project from the "agent" template with name "triagebot"
    Then the output directory should contain "main.go"
    And the output directory should contain "go.mod"
