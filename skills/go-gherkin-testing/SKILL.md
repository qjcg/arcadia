---
name: go-gherkin-testing
description: Add observable-behavior tests to Go projects using Gherkin feature files and godog (Cucumber for Go). Use when the user asks to add BDD tests, acceptance tests, integration tests, behavior-driven tests, Cucumber tests, Gherkin scenarios, godog, feature files, step definitions, scenario-based testing, or observable-behavior testing to a Go project. This covers the full workflow: adding the godog dependency, writing .feature files with Given/When/Then scenarios, organizing step definitions, wiring the godog test suite, and running tests. Also use when discussing testing strategy, test organization, or writing readable executable specifications in Go.
---

# Go Gherkin Testing

Add observable-behavior tests to Go projects using Gherkin feature files and
godog (Cucumber for Go). Tests are written as plain-language scenarios in
`.feature` files with step definitions in Go.

This skill covers the full workflow — from adding the dependency through
writing scenarios, organizing step definitions, and running the suite.

## Principles

Read these references before writing feature files or step definitions:

- **`references/better-gherkin.md`** — Write declarative scenarios that
  describe *what* the system does, not *how*. Hide implementation details in
  step definitions. If you read nothing else, read this.
- **`references/gherkin-spec.md`** — Full Gherkin syntax reference: keywords,
  Scenario Outlines, Doc Strings, Data Tables, Background, tags.
- **`references/step-organization.md`** — Organise step definitions by domain
  concept, not by feature. Use helper methods to avoid duplication.
- **`references/anti-patterns.md`** — Avoid feature-coupled steps and
  conjunction steps that hurt reusability.
- **`references/anti-patterns-blog.md`** — Deeper discussion: writing Gherkin
  *before* code, keeping scenarios concrete, avoiding UI details in steps.
- **`references/who-does-what.md`** — The Three Amigos collaboration model:
  who writes features, who writes step definitions, and how the team
  discovers examples together. Read this to understand the BDD process.
- **`references/user-story.md`** — User story format (As a / I want / So that)
  and how acceptance criteria map to Gherkin scenarios. Read this when a user
  story already exists and you need to translate it into a feature file.
- **`references/godog-examples.md`** — Index of 7 complete, runnable godog
  example projects (godogs, api, db, etc.). Read this when you need to see a
  real working project from feature files through step definitions to test
  suite wiring.

## Workflow

### 1. Collaborate on scenarios (Three Amigos)

Before writing any code or Gherkin, the team must discover concrete examples
*together*. This is the heart of BDD (from who-does-what.md).

The **Three Amigos** meeting brings together three perspectives:

- **Product owner** — defines scope: what user stories are in or out
- **Tester** — generates scenarios and edge cases: how might the application break?
- **Developer** — adds technical detail: what are the roadblocks and hidden requirements?

Each amigo sees the product from a different angle. Scenarios discovered
through this collaboration are more thorough, more realistic, and less likely
to miss edge cases than any one person writing alone.

The output of this step is a shared understanding of the examples you will
codify in feature files. If a user story already exists (see
`references/user-story.md`), use it as the starting point for the
conversation — the `<actor>` and `<feature>` map directly to the
feature file narrative section.

### 2. Add godog dependency

```bash
go get github.com/cucumber/godog@v0.15.1
```

Check `go.mod` first to see if it is already present.

**Study the minimal example** — `references/godog-examples/godogs/` is a
complete, runnable godog project in ~50 lines. Read it to see how feature
files, step definitions, and the test suite fit together before building your
own.

### 3. Identify observable behaviours

For each behaviour to test, ask: *what does a user or external caller observe?*

| System type  | Observable outcomes                              |
|--------------|--------------------------------------------------|
| CLI tool     | stdout/stderr text, exit codes, created files    |
| HTTP server  | response status codes, bodies, headers           |
| Library      | return values, side effects (filesystem, etc.)   |
| TUI          | rendered output, screen state transitions        |

Resist checking internal state (database records, in-memory data). Test only
what a user would observe.

### 4. Write feature files (Given/When/Then)

Place `.feature` files in a `features/` directory at the project root.

**Structural rules** (from the Gherkin spec):
- One `Feature` per `.feature` file
- One scenario per observable behaviour
- 3-5 steps per scenario (too many steps loses expressive power)
- Use `And` / `But` instead of repeating `Given` / `Then`
- Use `Background` to extract repeated `Given` steps — but keep it short (<4 lines)
- See [godog's `outline.feature`](https://github.com/cucumber/godog/blob/26931e66028d28bc7522af082fb55f1d57628ceb/features/outline.feature)
  for a real-world `Scenario Outline` with `Examples` tables
- See [godog's `tags.feature`](https://github.com/cucumber/godog/blob/26931e66028d28bc7522af082fb55f1d57628ceb/features/tags.feature)
  for tag usage patterns

**Feature narrative section** — Connect the feature back to its user story
(from user-story.md). The narrative is the free-form text between the
`Feature:` line and the first scenario. It is not executed, but it is
available for reporting and documentation. Use it to capture **Why**,
**Who**, and **What**:

```gherkin
Feature: Account balance
  In order to make better informed decisions about my spending
  As a mobile bank customer
  I want to see the balance on my accounts
```

- `In order to` — the reason/justification (protect revenue, increase value, etc.)
- `As a` — the role being served
- `I want` — one-sentence explanation of the feature

Keep the narrative short and meaningful. Avoid mechanical fill-ins like
"As a user, I want to check my balance, so I know what my balance is" —
that adds no information. Instead, describe the business rule or
uncertainty that the scenarios will illustrate.

**Naming conventions:**
- Feature files: `kebab-case.feature` (e.g., `list-templates.feature`)
- Scenario names: descriptive sentence case — think *"the one where..."*
  - Good: `Balance check with insufficient funds`
  - Bad: `Sign up, login, go to balance screen, check balance, logout`
- Step text: present tense, use quotation marks for arguments

```gherkin
Feature: Built-in templates
  The tool lists and hydrates project templates.

  Scenario: List shows all templates
    When I run the app with "-l"
    Then the output should contain "web"
    And the output should contain "library"

  Scenario: Hydrate a template
    Given the "web" template exists
    When I hydrate the "web" template with name "my-site"
    Then the output directory should contain "index.html"
```

**Error handling** gets its own feature file with multiple scenarios:

```gherkin
Feature: Error handling
  The tool reports clear errors for invalid usage.

  Scenario: Non-existent template
    When I run the app with "-t", "nonexistent"
    Then the output should contain "not found"

  Scenario: Output directory exists
    Given an existing non-empty output directory
    When I hydrate a template into that directory
    Then the output should contain "exists and is not empty"
```

**See a real working example** — `references/godog-examples/godogs/features/godogs.feature`
is a minimal 12-line feature file. Read it alongside the corresponding
`godogs_test.go` to see how feature text maps to step definitions.

**HTTP feature file pattern** — For a REST API example, see
`references/godog-examples/api/features/version.feature` which tests a
version endpoint with JSON responses.

**Use `Rule` to group scenarios by business rule** (Gherkin v6+).

A `Feature` can contain multiple `Rule` sections, each collecting the
scenarios that illustrate one business rule. This turns feature files into
structured living documentation — readers see the rule's abstract description
followed by concrete examples.

```gherkin
Feature: Account withdrawals

  Rule: A customer cannot withdraw more than their current balance

    Scenario: Balance check with sufficient funds
      Given a current balance of £100
      When I withdraw £50
      Then the withdrawal should succeed
      And the new balance should be £50

    Scenario: Balance check with insufficient funds
      Given a current balance of £30
      When I withdraw £50
      Then the withdrawal should be rejected
      And the balance should remain £30

  Rule: ATM withdrawals are capped at £200 per transaction

    Scenario: Withdrawal within limit
      Given a current balance of £500
      When I withdraw £150
      Then the withdrawal should succeed

    Scenario: Withdrawal exceeds limit
      Given a current balance of £500
      When I withdraw £250
      Then the withdrawal should be rejected
```

**When to use `Rule`:**
- The feature has multiple business rules that each need several example scenarios
- You want the `.feature` file to double as readable business documentation
- You need `Background` at a sub-feature level (each `Rule` can have its own `Background`)

**When to skip `Rule`:**
- Simple features with 2-3 scenarios and one implicit rule
- The scenarios already read clearly without the extra nesting level

**Declarative over imperative** (from better-gherkin.md):

```gherkin
# Good — declarative, describes behaviour
Scenario: Free subscribers see only free articles
  Given Free Frieda has a free subscription
  When Free Frieda logs in
  Then she sees only free articles

# Avoid — imperative, describes UI mechanics
Scenario: Free subscribers see only free articles
  Given I visit "/login"
  When I enter "freeFrieda@example.com" in the "email" field
  And I enter "password123" in the "password" field
  And I press the "Submit" button
  Then I see "FreeArticle1" on the home page
```

### 5. Create step definitions

Create step definitions in `features/steps/`. **Organise by domain concept,
not by feature** (from step-organization.md). A file per domain object or
capability avoids the feature-coupled anti-pattern.

```
features/steps/
  authentication_steps.go
  template_steps.go
  error_steps.go
```

**See complete examples** — `references/godog-examples/` has 7 working
projects. Start with `godogs/` for the minimal setup, then `api/` for HTTP
patterns, and `db/` for database state management. Each directory is a
standalone Go module with its own feature files and step definitions.

All scenarios share a single state struct. Reset it before each scenario.

```go
package steps

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var stepsDir string
var projectRoot string
var FeaturesDir string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		stepsDir = filepath.Dir(filename)
		FeaturesDir = filepath.Dir(stepsDir)
		projectRoot = filepath.Dir(FeaturesDir)
	}
}

type State struct {
	binPath      string
	tmpDir       string
	lastOutput   string
	lastExitCode int
	outputDir    string
}
```

**Why `runtime.Caller(0)` in `init()`**: `go test` may run from a different
working directory relative to the feature files. Do NOT compute paths in a
`var` initializer — those run before `init()` and yield empty paths.

#### Helper methods

```go
func (s *State) buildBinary() error {
	if s.binPath != "" {
		return nil
	}
	s.binPath = filepath.Join(s.tmpDir, "myapp")
	cmd := exec.Command("go", "build", "-o", s.binPath, ".")
	cmd.Dir = projectRoot
	cmd.Env = append(os.Environ(), "GOWORK=off")
	_, err := cmd.CombinedOutput()
	return err
}

func (s *State) runApp(args ...string) (string, error) {
	cmd := exec.Command(s.binPath, args...)
	out, err := cmd.CombinedOutput()
	output := string(out)
	s.lastOutput = output
	s.lastExitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.lastExitCode = exitErr.ExitCode()
		} else {
			return output, err
		}
	}
	return output, nil
}

func (s *State) fileExists(path string) bool {
	fullPath := filepath.Join(s.outputDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

func (s *State) reset() {
	s.lastOutput = ""
	s.lastExitCode = 0
	s.outputDir = filepath.Join(s.tmpDir, "output")
	os.RemoveAll(s.outputDir)
	os.MkdirAll(s.outputDir, 0755)
}
```

#### Step registration

Group related steps into registration functions. Godog deduplicates by
pattern text, so shared patterns (e.g., `the output should contain "..."`)
can be registered from multiple groups without conflict.

```go
package steps

import "github.com/cucumber/godog"

func RegisterListSteps(ctx *godog.ScenarioContext, s *State) {
	ctx.Step(`^I run the app with "([^"]+)"$`, func(flags string) error {
		return s.runWithFlags(flags)
	})
	ctx.Step(`^the output should contain "([^"]+)"$`, func(expected string) error {
		return s.outputContains(expected)
	})
}

func RegisterTemplateSteps(ctx *godog.ScenarioContext, s *State) {
	ctx.Step(`^the "([^"]+)" template exists$`, func(name string) error {
		return s.templateExists(name)
	})
	ctx.Step(`^I hydrate the "([^"]+)" template with name "([^"]+)"$`,
		func(template, name string) error {
			return s.hydrateTemplate(template, name)
		})
}
```

#### Assertion helpers

Keep these in a shared file. Return `nil` on pass, error on failure (godog
treats non-nil errors as step failures).

```go
func containsStr(s, substr string) error {
	if strings.Contains(s, substr) {
		return nil
	}
	return fmt.Errorf("expected output to contain %q, got:\n%s", substr, s)
}

func errf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
```

### 6. Wire the godog test suite

Create `features/steps/myapp_test.go` as the test entry point:

```go
package steps

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	state := &State{tmpDir: t.TempDir()}
	if err := state.buildBinary(); err != nil {
		t.Fatalf("building binary: %v", err)
	}

	suite := godog.TestSuite{
		Name: "myproject",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
				state.reset()
				return ctx, nil
			})
			RegisterListSteps(ctx, state)
			RegisterTemplateSteps(ctx, state)
			// ... register all step groups
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{FeaturesDir},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status from godog test suite")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
```

**Key design decisions:**
- Build the binary **once per test suite** (in `TestFeatures`), not per scenario
- Reset state in `ctx.Before` so each scenario starts fresh
- Use `cmd.Dir` to run `go build` from the project root when the test package is in a subdirectory
- **Non-zero exit codes are not errors from godog's perspective** — capture the exit code in state, then assert on it in subsequent `Then` steps. Do NOT return an error from the `When` step when a command fails.

### 7. Create test data (fixtures)

For scenarios that need input files, place them in `features/steps/testdata/`:

```
features/steps/testdata/
  custom/
    config.yaml
    main.go.tmpl
```

Reference testdata paths using `stepsDir` (computed from `runtime.Caller(0)`):

```go
func (s *State) setupFixture() error {
	s.customDir = filepath.Join(stepsDir, "testdata", "custom")
	_, err := os.Stat(s.customDir)
	return err
}
```

### 8. Run tests

```bash
go test ./features/steps/
go test -v ./features/steps/   # verbose: see which scenarios pass
go test ./...                   # full suite
```

## Anti-patterns to avoid

These are drawn from the reference material. Violating them leads to brittle,
hard-to-maintain test suites.

| Anti-pattern                                                                                 | Why it hurts                                             | Fix                                                                  |
|----------------------------------------------------------------------------------------------|----------------------------------------------------------|----------------------------------------------------------------------|
| **Feature-coupled steps** — step defs named after features (`edit_work_experience_steps.go`) | Duplication, can't reuse steps across features           | Group by domain concept (`employee_steps.go`)                        |
| **Conjunction steps** — `Given I have shades and a brand new Mustang`                        | Steps too specialised, hard to reuse                     | Split: `Given I have shades` / `And I have a brand new Mustang`      |
| **UI details in steps** — `Given I click the login button`                                   | Brittle, changes whenever UI changes, poor documentation | Declarative: `Given "Bob" logs in`                                   |
| **Imperative style** — describes keystrokes instead of intent                                | Obscures business rule, high maintenance                 | Declarative: describes *what*, not *how*                             |
| **Multiple `When` events** — two different actions in the same scenario                      | Tests multiple behaviours at once                        | Split into separate scenarios                                        |
| **Scenario Outline overuse** — parameterised everything                                      | Tests become slow and unfocused                          | Use outlines only for algorithmic/data-driven cases                  |
| **`Given`/`When`/`Then` boundary blur** — using `When` for setup                             | Confuses readers about intention                         | `Given` = past / context, `When` = action / event, `Then` = outcome  |
| **Incidental details** — password values, URLs in step text                                  | Obscures what the scenario actually tests                | Move into step definitions, keep feature text focused on the rule    |
| **All scenarios kept forever**                                                               | Clutter, stale documentation                             | Delete or rewrite scenarios that no longer test meaningful behaviour |
| **Writing Gherkin after code**                                                               | Misses the collaboration benefit of BDD                  | Write scenarios *before* implementing (the spec drives the code)     |
| **Boring scenarios** — `Given my account is empty / When I check / Then it is 0`             | Tests obvious behaviour with no new insight              | Replace with edge cases that document real business rules            |
| **No `Rule` when feature has 6+ scenarios across multiple rules**                            | Flat list of scenarios hides the business rule structure | Group by business rule with `Rule` sections to make document structure visible |
| **Feature files in the wrong location** — `.feature` files outside `features/`               | godog can't find them, tests fail with cryptic errors     | Always place `.feature` files in `features/` at the project root |

The godog repository includes a `references/godog-examples/incorrect-project-structure/`
example that deliberately demonstrates this last anti-pattern — study it to
understand how godog discovers feature files.

## Step definition patterns

### CLI tool steps

```go
func (s *State) runWithFlags(flags string) error {
	if err := s.buildBinary(); err != nil {
		return err
	}
	args := parseArgs(flags)
	_, err := s.runApp(args...)
	return err
}

func (s *State) outputContains(expected string) error {
	return containsStr(s.lastOutput, expected)
}
```

### HTTP server steps

**See the complete API example** — `references/godog-examples/api/` is a
dedicated HTTP testing project with its own README. It shows how to start a
`httptest` server, send requests, and assert on JSON responses.

```go
func (s *State) sendRequest(method, path string) error {
	req, _ := http.NewRequest(method, s.serverURL+path, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	s.lastResponse = resp
	s.lastBody, _ = io.ReadAll(resp.Body)
	return nil
}

func (s *State) responseCodeIs(expected int) error {
	if s.lastResponse.StatusCode != expected {
		return errf("expected status %d, got %d", expected, s.lastResponse.StatusCode)
	}
	return nil
}
```

### File system steps

```go
func (s *State) outputDirContains(path string) error {
	if !s.fileExists(path) {
		return errf("expected file %s to exist in output directory", path)
	}
	return nil
}

func (s *State) fileContains(path, expected string) error {
	data, err := os.ReadFile(filepath.Join(s.outputDir, path))
	if err != nil {
		return err
	}
	return containsStr(string(data), expected)
}
```

## Scenario quality checklist

Before writing a new scenario, verify:

1. **Is it declarative?** Does it describe *what* the system should do, not *how*?
2. **Is it observable?** Does the `Then` check something a user would see?
3. **Is it concrete?** Are the values specific (not "some money" but "£50")?
4. **Is it focused?** Does it test exactly one behaviour?
5. **Is it brief?** Can the reader understand it at a glance? (3-5 steps)
6. **Is the name descriptive?** Would someone unfamiliar with the domain
   understand what this scenario validates?
7. **Are incidental details absent?** If a value exists only because the
   automation needs it (passwords, IDs), move it to the step definition.
