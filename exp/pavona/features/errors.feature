Feature: Error handling
  Pavona reports clear errors for invalid usage.

  Scenario: Non-existent template name
    When I run pavona with "-t", "nonexistent"
    Then the output should contain "not found"

  Scenario: Output directory exists and is not empty
    Given an existing non-empty output directory
    When I hydrate the "tool" template into that directory
    Then the output should contain "exists and is not empty"
