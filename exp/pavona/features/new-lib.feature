Feature: Lib template
  The lib template generates a minimal Go library module.

  Scenario: Create lib project
    When I create a project from the "lib" template with name "go-csvstream"
    Then the output directory should contain:
      | lib.go       |
      | lib_test.go  |
      | go.mod       |
      | Taskfile.yaml |
