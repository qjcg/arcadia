Feature: Go monorepo template
  The monorepo-go template generates a Go workspace with repository tooling and release automation.

  Scenario: Create monorepo-go project
    When I create a project from the "monorepo-go" template with name "my-monorepo"
    Then the output directory should contain "go.work"
    And the output directory should contain "Taskfile.yaml"
    And the output directory should contain "README.md"
    And the output directory should contain "AGENTS.md"
    And the output directory should contain ".editorconfig"
    And the output directory should contain ".editorconfig-checker.json"
    And the output directory should contain ".lefthook.yaml"
    And the output directory should contain ".github/CODEOWNERS"
    And the output directory should contain ".github/workflows/sv-release.yml"
    And the output directory should contain "docs"
    And "go.work" should contain "go 1.27.1"
    And "Taskfile.yaml" should contain "TODO build"
    And ".github/workflows/sv-release.yml" should contain "actions/checkout"
