Feature: Worker pool — pool package
  The pool package provides background job execution.

  Scenario: Pool runs jobs concurrently
    When I submit 10 jobs to a pool of 4 workers
    Then at most 4 jobs run concurrently

  Scenario: Pool drains on shutdown
    Given 3 running jobs in the pool
    When the app stops
    Then the pool waits for running jobs to finish
    And does not accept new jobs
