Feature: Core runtime — app lifecycle
  The app package manages module registration and lifecycle.

  Scenario: Create an app with no modules
    When I create a pavona.App with no modules
    Then app.Start() does not error
    And app.Stop() does not error

  Scenario: Modules start in registration order
    Given modules A and B where A depends on B
    When the app starts
    Then B starts before A
    And A stops before B

  Scenario: Module health checks are collected
    Given a module that reports healthy
    And a module that reports unhealthy
    When I query app.Health()
    Then it returns one healthy and one unhealthy result

  Scenario: App shuts down gracefully on signal
    When the app receives SIGINT
    Then all modules receive stop before the process exits
    And in-flight work completes within the deadline

  Scenario: Config loads from layered sources
    When config is provided via file, env, and flags
    Then flags override env, env overrides file, file overrides defaults
