Feature: Static site builder — site package
  The site package builds Markdown and org-mode content into static HTML.

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
