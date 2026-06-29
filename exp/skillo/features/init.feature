Feature: Initialize workspace
  The init command sets up a Go modules workspace for managing skill versions.

  Background:
    Given a clean home directory

  Scenario: Init creates a go.mod file
    When I run "skillo" with "init"
    Then it should succeed
    And the file ".skillo/go.mod" should exist
    And the file ".skillo/go.mod" should contain "module skillo.local/skills"

  Scenario: Re-init reports already initialized
    When I run "skillo" with "init"
    And I run "skillo" with "init"
    Then it should succeed
    And the output should contain "workspace already initialized"

  Scenario: Init with custom modules dir
    When I run "skillo" with "init --modules-dir " and the path "custom-skillo"
    Then it should succeed
    And the file "custom-skillo/go.mod" should exist
