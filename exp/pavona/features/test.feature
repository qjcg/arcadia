Feature: Test helpers — test package
  The test package provides shared helpers and godog step definitions.

  Scenario: Temp database for tests
    When a test requests a temp database
    Then it gets a clean SQLite database
    And migrations are run automatically
    And the database is cleaned up after the test

  Scenario: Golden file comparison
    When a test produces output
    Then it can compare against a golden file
    And the comparison detects differences

  Scenario: Request recording
    When a test sends an HTTP request
    Then the request is recorded with method, path, headers, body
    And the recorded requests are assertable

  Scenario: Clock mocking
    When a test uses the mocked clock
    Then time does not advance unless explicitly advanced
    And timeouts can be triggered deterministically

  Scenario: Fresh scaffold has passing features
    When I scaffold any project type
    Then `godog run features/` passes with zero failures

  Scenario: BDD steps are provided by the test package
    Given a scaffolded project
    Then the godog steps for common operations exist
    And custom steps can be added per project
