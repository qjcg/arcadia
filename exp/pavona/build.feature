Feature: Pavona CLI — project scaffolding
  The pavona CLI creates, extends, and builds projects of all six types.

  Scenario: Scaffold a CLI tool
    When I run `pavona new tool gh-deploy`
    Then a directory "gh-deploy" exists
    And the file "gh-deploy/main.go" exists
    And the file "gh-deploy/features/" exists
    And `go build ./...` succeeds in "gh-deploy"

  Scenario: Scaffold a library
    When I run `pavona new lib go-csvstream`
    Then a directory "go-csvstream" exists
    And the library compiles with zero Pavona runtime dependency

  Scenario: Scaffold a static site with markdown
    When I run `pavona new site blog`
    Then a directory "blog" exists
    And the directory "blog/content/" exists
    And `pavona build` in "blog" produces "blog/dist/index.html"

  Scenario: Scaffold a static site with org-mode
    When I run `pavona new site blog --format org`
    Then "blog/content/" contains files with ".org" extension

  Scenario: Scaffold a TUI app
    When I run `pavona new tui chatmonitor`
    Then the TUI binary builds and responds to keyboard input

  Scenario: Scaffold a full-stack web app
    When I run `pavona new app acmecorp`
    Then the server binary builds and serves HTTP on startup
    And the server serves rendered templ templates

  Scenario: Scaffold an agent
    When I run `pavona new agent triagebot`
    Then the agent binary builds
    And it registers as a NATS micro service named "agents"

  Scenario: Add a handler to an existing app
    Given a pavona app project
    When I run `pavona add handler auth/login`
    Then a handler file exists at "handlers/auth/login.go"
    And the route is registered in the server

  Scenario: Add a BDD feature to an existing project
    Given any pavona project
    When I run `pavona add feature prompt_response`
    Then "features/prompt_response.feature" exists
    And "features/steps/prompt_response.go" exists
    And `godog run features/` passes

  Scenario: Remove an added component
    Given a project with a handler "auth/login"
    When I run `pavona remove handler auth/login`
    Then the handler file does not exist
    And the route is no longer registered


Feature: Pavona CLI — templates
  Templates overlay technology stacks on project types.

  Scenario: Scaffold with NATS template
    When I run `pavona new app dashboard --template nats`
    Then the project includes "nats/conn.go"
    And the embedded NATS server starts on app boot
    And the NATS server is configured via config.yaml

  Scenario: NATS integrates with the app kernel lifecycle
    Given a project scaffolded with `--template nats`
    When the app starts
    Then the embedded NATS server starts before modules
    And it stops gracefully after modules drain

  Scenario: JetStream stream and consumer definitions
    Given a project with the NATS template
    When I run `pavona add stream events`
    Then "nats/streams/events.go" exists
    And it declares a JetStream stream with a durable consumer

  Scenario: Create a custom template
    When I run `pavona template new my-stack`
    Then "templates/my-stack/template.yaml" exists
    And the template can be applied via `--template my-stack`

  Scenario: Publish and discover templates
    When I run `pavona template publish my-stack`
    Then the template is available in the registry
    When I run `pavona template search postgres`
    Then I see a list of matching templates


Feature: Core runtime — app lifecycle
  The pavona/app package manages module registration and lifecycle.

  Scenario: Create an app with no modules
    When I create a pavona.App with no modules
    Then app.Start() does not error
    And app.Stop() does not error

  Scenario: Modules start in registration order
    Given modules A and B where A depends on B
    When the app starts
    Then B starts before A
    And A stops before B

  Scenario: Module health checks are collected
    Given a module that reports healthy
    And a module that reports unhealthy
    When I query app.Health()
    Then it returns one healthy and one unhealthy result

  Scenario: App shuts down gracefully on signal
    When the app receives SIGINT
    Then all modules receive stop before the process exits
    And in-flight work completes within the deadline

  Scenario: Config loads from layered sources
    When config is provided via file, env, and flags
    Then flags override env, env overrides file, file overrides defaults


Feature: HTTP server — pavona/serve
  The serve package provides a standard net/http server with middleware.

  Scenario: Server starts and serves requests
    When I create a server with one route "GET /health"
    Then the server responds 200 on GET /health

  Scenario: Middleware stack applies in order
    Given middleware A (request-id) and B (access-log)
    When a request arrives
    Then A runs first, then B, then the handler

  Scenario: Built-in middleware is functional
    When a handler panics
    Then the recover middleware catches it and returns 500
    When a request has no request-id header
    Then the request-id middleware sets one

  Scenario: Server shuts down gracefully
    Given an in-flight request taking 5s
    When the server receives a stop signal
    Then it waits up to the configured deadline then exits


Feature: CLI framework — pavona/cli
  The cli package provides command routing and flag parsing.

  Scenario: Define a command with subcommands
    When I define a "deploy" command with subcommands "push" and "rollback"
    Then running "tool push --env staging" calls the push handler with env=staging

  Scenario: Default help is generated
    When I run a tool with no arguments
    Then it prints usage listing all subcommands

  Scenario: Config binds to flags
    Given a config field "log.level" mapped to "--log-level"
    When I run with `--log-level debug`
    Then the config value is "debug"


Feature: TUI framework — pavona/tui
  The tui package wraps bubbletea with layout and keybinding primitives.

  Scenario: A basic TUI renders and responds to keys
    When I run a TUI app
    Then I see a rendered screen
    And pressing "q" exits

  Scenario: Keybinding registry maps keys to actions
    When I register "ctrl+c" → quit and "enter" → submit
    Then pressing ctrl+c exits
    And pressing enter triggers submit

  Scenario: Help overlay is available
    When I press "?"
    Then a help overlay shows all registered keybindings


Feature: Static site builder — pavona/site
  The site package builds Markdown/org-mode content into static HTML.

  Scenario: Build produces HTML from markdown
    Given "content/index.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/index.html" exists and contains the rendered body

  Scenario: Org-mode content is rendered
    Given "content/index.org" with org headers and body
    When I run `pavona build`
    Then "dist/index.html" exists and contains the rendered content

  Scenario: Dev server watches for changes
    When I run `pavona serve`
    Then the dev server starts on localhost
    And changing a content file triggers a rebuild

  Scenario: Tailwind CSS is compiled
    When I run `pavona build`
    Then "dist/styles.css" exists with compiled Tailwind output


Feature: Database — pavona/db
  The db package provides migration runner and connection lifecycle.

  Scenario: Migration runs up
    Given a migration file "001_create_users.up.sql"
    When the app starts
    Then the "users" table exists in the database

  Scenario: Migration runs down
    Given the "users" table exists
    When I run `pavona db migrate down`
    Then the "users" table no longer exists

  Scenario: Migration status reports current version
    When I run `pavona db migrate status`
    Then it shows the current migration version and pending migrations


Feature: Agent protocol — pavona/agent
  The agent package implements the NATS Agent Protocol.

  Scenario: Agent registers as a NATS micro service
    When an agent starts
    Then `nats req '$SRV.INFO.agents'` lists the agent by name

  Scenario: Agent responds to prompt requests
    Given an agent with a prompt handler
    When I send a prompt request with text "hello"
    Then the response contains typed JSON chunks
    And the final chunk has empty body (terminator)

  Scenario: Agent reports status
    When I query the status endpoint
    Then I receive healthy=true and current load

  Scenario: Agent heartbeat is emitted
    Given a running agent
    Then a heartbeat message is published at the configured interval

  Scenario: Embedded NATS server for development
    When an agent starts with no external NATS server configured
    Then it embeds a memory-backed NATS server automatically


Feature: Worker pool — pavona/pool
  The pool package provides background job execution.

  Scenario: Pool runs jobs concurrently
    When I submit 10 jobs to a pool of 4 workers
    Then at most 4 jobs run concurrently

  Scenario: Pool drains on shutdown
    Given 3 running jobs in the pool
    When the app stops
    Then the pool waits for running jobs to finish
    And does not accept new jobs


Feature: Templates — pavona/gen
  The gen package provides the code generation and template engine.

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


Feature: Testing — pavona/test
  The test package provides shared test helpers and godog steps.

  Scenario: Temp database for tests
    When a test requests a temp database
    Then it gets a clean SQLite database
    And migrations are run automatically
    And the database is cleaned up after the test

  Scenario: Golden file comparison
    When a test produces output
    Then it can compare against a golden file
    And the comparison detects differences

  Scenario: Request recording
    When a test sends an HTTP request
    Then the request is recorded with method, path, headers, body
    And the recorded requests are assertable

  Scenario: Clock mocking
    When a test uses the mocked clock
    Then time does not advance unless explicitly advanced
    And timeouts can be triggered deterministically


Feature: Generated project — gherkin features pass
  Every scaffolded project includes a running godog suite.

  Scenario: Fresh scaffold has passing features
    When I scaffold any project type
    Then `godog run features/` passes with zero failures

  Scenario: BDD steps are provided by pavona/test
    Given a scaffolded project
    Then the godog steps for common operations exist
    And custom steps can be added per project
