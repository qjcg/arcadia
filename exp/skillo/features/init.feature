Feature: Initialize workspace
  The init command sets up a skillo workspace with go.mod and selections.json.

  Background:
    Given a clean home directory

  Scenario: Init creates a go.mod file
    When I run "skillo" with "init"
    Then it should succeed
    And the file ".config/skillo/go.mod" should exist
    And the file ".config/skillo/go.mod" should contain "module skillo.local/skills"

  Scenario: Re-init reports already initialized
    When I run "skillo" with "init"
    And I run "skillo" with "init"
    Then it should succeed
    And the output should contain "workspace already initialized"

  Scenario: Init with --project flag creates project .skillo/
    When I run "skillo" with "init --project"
    Then it should succeed
