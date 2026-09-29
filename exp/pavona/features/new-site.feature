Feature: Site template
  The site template generates a static site project.

  Scenario: Create site project
    When I create a project from the "site" template with name "blog"
    Then the output directory should contain "content/index.md"
