Feature: HTTP server — serve package
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
