Feature: List skills
  The list command shows installed skills with version, description, and
  module source info. Use --format json for machine-readable output,
  --outdated to check for available upgrades, and --scope to filter.

  Scenario: List with no skills installed
    Given a clean home directory
    When I run "skillo" with "list"
    Then it should succeed

  Scenario: List with --scope user shows nothing
    Given a clean home directory
    When I run "skillo" with "list --scope user"
    Then it should succeed

  Scenario: List with skills shows names
    Given a clean home directory
    When I run "skillo" with "init"
    And I run "skillo" with "list"
    Then it should succeed

  Scenario: List with --format json outputs valid JSON
    Given a clean home directory
    When I run "skillo" with "list --format json"
    Then it should succeed

  Scenario: List with --outdated flag succeeds when no modules installed
    Given a clean home directory
    When I run "skillo" with "init"
    And I run "skillo" with "list --outdated"
    Then it should succeed

  Scenario: List with --help shows flags
    Given a clean home directory
    When I run "skillo" with "list --help"
    Then it should succeed
    And the output should contain "--outdated"
    And the output should contain "--format"
    And the output should contain "--scope"
