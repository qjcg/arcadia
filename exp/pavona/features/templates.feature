Feature: NATS templates
  Templates overlay technology stacks on project types.

  Scenario: Scaffold with NATS template
    When I scaffold an "app" named "dashboard"
    And the scaffold includes "nats/conn.go" and NATS dependencies

  Scenario: NATS integrates with the app kernel lifecycle
    Given a project scaffolded with the NATS template
    When the app starts
    Then the embedded NATS server starts before modules
    And it stops gracefully after modules drain

  Scenario: JetStream stream definitions
    Given a project with the NATS template
    When I run `pavona add stream events`
    Then "nats/streams/events.go" exists with a stream and consumer

  Scenario: Create a custom template
    When I run `pavona template new my-stack`
    Then "templates/my-stack/template.yaml" exists
    And the template applies via `--template my-stack`
