Feature: Site template
  The site template generates a static site project.

  Scenario: Hydrate site template
    When I hydrate the "site" template with name "blog"
    Then the output directory should contain "content/index.md"
