Feature: Go monorepo template
  The monorepo-go template generates a Go workspace with repository tooling and release automation.

  Scenario: Create monorepo-go project
    When I create a project from the "monorepo-go" template with name "my-monorepo"
    Then the output directory should contain:
      | go.work                          |
      | Taskfile.yaml                    |
      | README.md                        |
      | AGENTS.md                        |
      | .editorconfig                    |
      | .editorconfig-checker.json       |
      | .lefthook.yaml                   |
      | .github/CODEOWNERS               |
      | .github/workflows/sv-release.yml |
      | docs                             |
    And "go.work" should contain "go 1.27.1"
    And "Taskfile.yaml" should contain "TODO build"
    And ".github/workflows/sv-release.yml" should contain "actions/checkout"
