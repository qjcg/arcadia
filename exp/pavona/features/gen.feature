Feature: Code generation — gen package
  The gen package provides the template engine and scaffolding.

  Scenario: Template applies files to a project
    When I scaffold a project with `--template my-stack`
    Then all files declared in the template are created
    And the template's go.mod dependencies are added

  Scenario: Template hooks run after scaffolding
    Given a template with a post-scaffold hook
    When the project is scaffolded
    Then the hook runs and modifies a generated file

  Scenario: Template overrides project type files
    Given a template that overrides "main.go"
    When the project is scaffolded
    Then "main.go" contains the template's version, not the default
