Feature: App template
  The app template generates a full-stack web app.

  Scenario: Hydrate app template
    When I hydrate the "app" template with name "acmecorp"
    Then the output directory should contain "main.go"
    And the output directory should contain "main_test.go"
    And the output directory should contain "go.mod"
    And the output directory should contain "Dockerfile"
    And the output directory should contain "internal/handlers/health.go"
