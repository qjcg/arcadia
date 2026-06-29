Feature: Lib template
  The lib template generates a minimal Go library module.

  Scenario: Hydrate lib template
    When I hydrate the "lib" template with name "go-csvstream"
    Then the output directory should contain "lib.go"
    And the output directory should contain "lib_test.go"
    And the output directory should contain "go.mod"
    And the output directory should contain "Taskfile.yaml"
