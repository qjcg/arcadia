Feature: App template
  The app template generates a full-stack web app.

  Scenario: Create app project
    When I create a project from the "app" template with name "acmecorp"
    Then the output directory should contain:
      | main.go                     |
      | main_test.go                |
      | go.mod                      |
      | Dockerfile                  |
      | internal/handlers/health.go |
