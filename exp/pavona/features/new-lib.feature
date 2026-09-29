Feature: Lib template
  The lib template generates a minimal Go library module.

  Scenario: Create lib project
    When I create a project from the "lib" template with name "go-csvstream"
    * the output directory should contain "lib.go"
    * the output directory should contain "lib_test.go"
    * the output directory should contain "go.mod"
    * the output directory should contain "Taskfile.yaml"
