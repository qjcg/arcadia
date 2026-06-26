Feature: Static site builder — site package
  The site package builds Markdown and org-mode content into static HTML.

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
