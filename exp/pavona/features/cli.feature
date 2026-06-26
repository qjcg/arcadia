Feature: CLI Tool scaffolding
  The pavona CLI can scaffold a new CLI tool project.

  Scenario: Scaffold a CLI tool with default name
    When I scaffold a "tool" named "gh-deploy"
    Then the project "gh-deploy" should exist
    And "gh-deploy/main.go" should contain "gh-deploy"
    And "gh-deploy/go.mod" should exist
    And "gh-deploy/Taskfile.yaml" should exist
    And "gh-deploy/.gitignore" should exist
    And "gh-deploy/features/" should exist
    And the project should compile

  Scenario: Scaffold a library
    When I scaffold a "lib" named "go-csvstream"
    Then the project "go-csvstream" should exist

  Scenario: Scaffold a static site
    When I scaffold a "site" named "blog"
    Then the project "blog" should exist

  Scenario: Scaffold a TUI
    When I scaffold a "tui" named "chatmonitor"
    Then the project "chatmonitor" should exist

  Scenario: Scaffold a web app
    When I scaffold a "app" named "acmecorp"
    Then the project "acmecorp" should exist

  Scenario: Scaffold an agent
    When I scaffold a "agent" named "triagebot"
    Then the project "triagebot" should exist

  Scenario: Error on unknown project type
    When I scaffold a "foo" named "bar"
    Then I should get an error about unknown project type
