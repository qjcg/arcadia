Feature: Validate skills
  The validate command checks that a SKILL.md file has valid frontmatter.

  Scenario: Validate a valid skill directory
    Given a valid skill directory
    When I run "skillo" with "validate" and the path
    Then it should succeed
    And the output should contain "Extracted skill"

  Scenario: Validate a directory without SKILL.md
    Given an empty directory
    When I run "skillo" with "validate" and the path
    Then it should succeed
    And the output should contain "Warning: No SKILL.md found"
