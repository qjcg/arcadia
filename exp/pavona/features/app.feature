Feature: Full-stack web app scaffold — pavona new app
  The app scaffold produces a ready-to-run web project with a-h/templ,
  SQLite, HTMX, Alpine.js, and Tailwind/DaisyUI via CDN.

  Scenario: Scaffold creates the project directory
    When I scaffold an "app" named "acmecorp"
    Then the project "acmecorp" should exist

  Scenario: Core project files exist
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/main.go" should exist
    And "acmecorp/go.mod" should exist
    And "acmecorp/Taskfile.yaml" should exist
    And "acmecorp/Dockerfile" should exist
    And "acmecorp/.gitignore" should exist
    And "acmecorp/config.yaml" should exist

  Scenario: Internal directory structure
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/internal/handlers/health.go" should exist
    And "acmecorp/internal/views/layout.templ" should exist
    And "acmecorp/internal/views/index.templ" should exist
    And "acmecorp/internal/db/schema.sql" should exist
    And "acmecorp/internal/db/db.go" should exist
    And "acmecorp/internal/static/style.css" should exist

  Scenario: Main.go sets up routes and server
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/main.go" should contain "net/http"
    And "acmecorp/main.go" should contain "ListenAndServe"
    And "acmecorp/main.go" should contain "/health"
    And "acmecorp/main.go" should contain "static"
    And "acmecorp/main.go" should contain "internal/handlers"
    And "acmecorp/main.go" should contain "internal/views"

  Scenario: Templ components are used for rendering
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/internal/views/layout.templ" should contain "templ Layout"
    And "acmecorp/internal/views/index.templ" should contain "templ Index"

  Scenario: Layout includes HTMX and Alpine.js
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/internal/views/layout.templ" should contain "htmx.org"
    And "acmecorp/internal/views/layout.templ" should contain "alpinejs"

  Scenario: Layout uses Tailwind and DaisyUI from CDN
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/internal/views/layout.templ" should contain "cdn.tailwindcss.com"
    And "acmecorp/internal/views/layout.templ" should contain "daisyui"

  Scenario: Database setup with SQLite
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/internal/db/schema.sql" should contain "CREATE TABLE"
    And "acmecorp/internal/db/db.go" should contain "database/sql"
    And "acmecorp/internal/db/db.go" should contain "sqlite"
    And "acmecorp/go.mod" should contain "modernc.org/sqlite"
    And "acmecorp/sqlc.yaml" should exist

  Scenario: Health handler returns status
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/internal/handlers/health.go" should contain "status"

  Scenario: BDD test suite is scaffolded
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/features/health.feature" should exist
    And "acmecorp/features/steps/health.go" should exist
    And "acmecorp/features/health.feature" should contain "health"
    And "acmecorp/main_test.go" should exist
    And "acmecorp/main_test.go" should contain "godog"

  Scenario: Taskfile with common targets
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/Taskfile.yaml" should contain "build"
    And "acmecorp/Taskfile.yaml" should contain "dev"
    And "acmecorp/Taskfile.yaml" should contain "test"
    And "acmecorp/Taskfile.yaml" should contain "templ"

  Scenario: Dockerfile builds and exposes the app
    When I scaffold an "app" named "acmecorp"
    Then "acmecorp/Dockerfile" should contain "golang"
    And "acmecorp/Dockerfile" should contain "EXPOSE"

  Scenario: Scaffolded project compiles
    When I scaffold an "app" named "acmecorp"
    Then the project should compile
