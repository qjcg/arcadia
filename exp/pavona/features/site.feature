Feature: Static site builder — site package
  The site package builds Markdown and org-mode content into static HTML
  with templ-based themes and a filesystem-as-URL structure.

  Scenario: Build produces HTML from markdown
    Given "content/index.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/index.html" exists and contains the rendered body

  Scenario: Build produces HTML from org-mode
    Given "content/index.org" with org headers and body
    When I run `pavona build`
    Then "dist/index.html" exists and contains the rendered content
    And "dist/index.html" contains org-specific HTML

  Scenario: Org content with bold and emphasis renders correctly
    Given "content/formatting.org" with org bold and emphasis
    When I run `pavona build`
    Then "dist/formatting.html" contains org formatting
    And "dist/formatting.html" contains the text "<em>"

  Scenario: Mixed markdown and org both produce HTML
    Given "content/intro.md" with frontmatter and body
    And "content/setup.org" with org headers and body
    When I run `pavona build`
    Then "dist/intro.html" exists and contains the rendered body
    And "dist/setup.html" exists and contains the rendered content

  Scenario: Multiple markdown files each produce their own HTML
    Given "content/index.md" with frontmatter and body
    And "content/about.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/index.html" exists and contains the rendered body
    And "dist/about.html" exists and contains the rendered body

  Scenario: Section with index.md produces clean /section/ URL
    Given "content/services/index.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/services/index.html" exists and contains the rendered body

  Scenario: Section sub-pages are rendered under the section directory
    Given "content/services/index.md" with frontmatter and body
    And "content/services/consulting.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/services/index.html" exists and contains the rendered body
    And "dist/services/consulting.html" exists and contains the rendered body

  Scenario: Nested sections create nested output directories
    Given "content/services/index.md" with frontmatter and body
    And "content/services/design/web.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/services/index.html" exists
    And "dist/services/design/web.html" exists

  Scenario: Title from YAML frontmatter overrides heading detection
    Given "content/page.md" with YAML title "Custom Title"
    When I run `pavona build`
    Then "dist/page.html" contains "Custom Title"

  Scenario: Order field controls navigation position
    Given "content/alpha.md" with YAML order "2"
    And "content/beta.md" with YAML order "1"
    When I run `pavona build`
    Then the navigation lists beta before alpha

  Scenario: Draft pages are excluded from build
    Given "content/index.md" with frontmatter and body
    And "content/draft.md" with YAML draft "true"
    When I run `pavona build`
    Then "dist/index.html" exists
    And "dist/draft.html" does not exist

  Scenario: Flat files produce .html URLs, sections produce /index.html
    Given "content/about.md" with frontmatter and body
    And "content/team/index.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/about.html" exists
    And "dist/team/index.html" exists

  Scenario: Navigation tree reflects section hierarchy
    Given "content/index.md" with frontmatter and body
    And "content/about.md" with frontmatter and body
    And "content/services/consulting.md" with frontmatter and body
    When I run `pavona build`
    Then the navigation contains a top-level link to "about"
    And the navigation contains a section "services"

  Scenario: Build fails with clear error when content directory is missing
    When I run `pavona build`
    Then I should get an error about the content directory

  Scenario: Dev server starts
    Given "content/index.md" with frontmatter and body
    When I run `pavona serve`
    Then the dev server starts on localhost

  Scenario: Dev server serves built files
    Given "content/index.md" with frontmatter and body
    When I run `pavona serve`
    Then the dev server serves the built file over HTTP

  Scenario: Site scaffold includes build.go and go.mod
    When I scaffold a "site" named "mysite"
    Then "mysite/build.go" should exist
    And "mysite/go.mod" should exist
    And "mysite/go.mod" should contain "github.com/qjcg/arcadia/exp/pavona"

  Scenario: Site scaffold with --pages creates content files
    When I scaffold a "site" named "mysite" with pages "about,contact"
    Then "mysite/content/about.md" should exist
    And "mysite/content/contact.md" should exist

  Scenario: Site scaffold with --pages and brace expansion
    When I scaffold a "site" named "mysite" with pages "services/{foo,bar},clients/{bar,baz}"
    Then "mysite/content/services/foo.md" should exist
    And "mysite/content/services/bar.md" should exist
    And "mysite/content/clients/bar.md" should exist
    And "mysite/content/clients/baz.md" should exist

  Scenario: Site scaffold with --pages in org format
    When I scaffold a "site" named "mysite" with format "org" and pages "about"
    Then "mysite/content/about.org" should exist

  Scenario: Site scaffold includes a theme directory
    When I scaffold a "site" named "mysite"
    Then "mysite/theme/" should exist
    And "mysite/theme/default.templ" should exist

  Scenario: Default theme wraps content in a full HTML page
    Given "content/index.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/index.html" contains "<!DOCTYPE html>"
    And "dist/index.html" contains "<html"
    And "dist/index.html" contains "<head>"
    And "dist/index.html" contains "<body"

  Scenario: Default theme has responsive navigation
    Given "content/index.md" with frontmatter and body
    And "content/about.md" with frontmatter and body
    When I run `pavona build`
    Then "dist/index.html" contains a navigation element
    And "dist/about.html" contains a navigation element

  Scenario: Build compiles the templ theme and renders content
    Given "content/index.md" with frontmatter and body
    When I run `pavona build`
    Then the build invokes `templ generate` on the theme
    And "dist/index.html" exists and contains the rendered body

  Scenario: Custom --theme flag selects a different theme directory
    Given a custom theme in "themes/docs/default.templ"
    When I run `pavona build --theme ./themes/docs`
    Then "dist/index.html" exists and contains the rendered body
