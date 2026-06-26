Feature: Database — db package
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
