Feature: Search skills
  The search command finds skills by name in local directories.

  Scenario: Search with no skills installed
    Given a clean home directory
    When I run "skillo" with "search" and the term "pdf"
    Then the output should contain "Searching for 'pdf'"
    And the output should contain "GitHub search integration coming soon"

  Scenario: Search after init finds no skills
    Given a clean home directory
    When I run "skillo" with "init"
    And I run "skillo" with "search" and the term "agent"
    Then the output should contain "Searching for 'agent'"
