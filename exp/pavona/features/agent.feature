Feature: Agent protocol — agent package
  The agent package implements the NATS Agent Protocol.

  Scenario: Agent registers as a NATS micro service
    When an agent starts
    Then `nats req '$SRV.INFO.agents'` lists the agent by name

  Scenario: Agent responds to prompt requests
    Given an agent with a prompt handler
    When I send a prompt request with text "hello"
    Then the response contains typed JSON chunks
    And the final chunk has empty body (terminator)

  Scenario: Agent reports status
    When I query the status endpoint
    Then I receive healthy=true and current load

  Scenario: Agent heartbeat is emitted
    Given a running agent
    Then a heartbeat message is published at the configured interval

  Scenario: Embedded NATS server for development
    When an agent starts with no external NATS server configured
    Then it embeds a memory-backed NATS server automatically
