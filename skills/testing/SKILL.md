---
name: testing
description: Guidance for writing high-quality unit tests using Vladimir Khorikov's principles from Unit Testing. Use when (1) writing new tests, (2) refactoring existing tests, (3) reviewing test code, or (4) deciding what to test.
---

# Testing - Khorikov's Unit Testing Principles

## The Four Pillars

Evaluate tests against these four criteria:

1. **Protection against regressions** - How well does the test catch bugs?
2. **Resistance to refactoring** - Will the test survive implementation changes?
3. **Fast feedback** - How quickly does the test run?
4. **Maintainability** - How easy is the test to understand and modify?

Prioritize based on context. For legacy code, focus on regression protection. For new code, balance all four.

## What to Test

### Test Behavior, Not Implementation
- Verify observable outcomes, not internal state
- Tests should pass if behavior is correct, regardless of how it's implemented
- Ask: "Can a non-developer understand what this test verifies?"

### The Domain Matters
- Business-critical logic: thorough testing
- Glue code: minimal testing
- Third-party libraries: no testing

## Test Structure

### AAA Pattern
```
// Arrange: Set up dependencies and inputs
// Act: Execute the behavior
// Assert: Verify the outcome
```

### One Assertion Per Test (Prefer)
Multiple assertions are acceptable when they verify a single logical outcome.

### Test Naming
Describe the scenario and expected behavior:
- `should_return_empty_list_when_no_items_exist`
- `should_throw_error_when_user_not_found`

## Mocking Guidelines

### Only Mock What You Own
- Mock shared dependencies (databases, external services)
- Don't mock your own code under test
- Avoid mocking value objects or entities

### The Principle
> Mock roles, not objects. Tests should verify collaboration, not implementation.

## Test Ethics

### Delete Bad Tests
Remove tests that:
- Frequently produce false positives
- Are expensive to maintain
- Don't detect real bugs

### False Positives Are Worse Than No Tests
A test that breaks during refactoring (when nothing is wrong) erodes trust and gets disabled.

## Decision Framework

| Scenario    | Approach                                 |
|-------------|------------------------------------------|
| New feature | Write tests alongside implementation     |
| Bug fix     | Add regression test                      |
| Refactoring | Keep tests that verify behavior          |
| Legacy code | Add tests when understanding, not before |

## Code Review Checklist

- [ ] Does the test verify behavior, not implementation?
- [ ] Is the test name descriptive?
- [ ] Are dependencies minimal?
- [ ] Would this test survive a reasonable refactor?
- [ ] Does it provide clear value if it fails?
