# Error Messages

Write errors that help users understand and recover.

## Principles

### Rewrite Technical Errors
Raw error messages from libraries are for developers, not users. Rewrite them.

```go
// Bad: Technical and unhelpful
failed to connect to database: dial tcp 127.0.0.1:5432: connection refused

// Good: User understands the problem
Cannot connect to database. Is the database server running?
Hint: Try 'myapp db start' to start the database.
```

### Include Context
- What operation was attempted
- Why it failed
- How to fix it (when known)

```go
// Good error structure
fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
fmt.Fprintf(os.Stderr, "Hint: %s\n", getHint(err))
```

## Best Practices

### Exit Codes
| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Misuse / invalid arguments |
| 3 | Configuration error |
| 4 | Resource not found |
| 5 | Authentication/permission error |
| 126 | Command not executable |
| 127 | Command not found |

### Error Message Checklist
- [ ] Is the message human-readable?
- [ ] Does it explain what went wrong?
- [ ] Does it suggest how to fix it?
- [ ] Is it proportionate to the severity?
- [ ] Does it avoid blaming the user?

### Signal-to-Noise Ratio
- Only show errors that matter to the user
- Suppress verbose debug info unless `--debug`
- Don't repeat the same error multiple times

## Table of Contents

- [Principles](#principles)
  - [Rewrite Technical Errors](#rewrite-technical-errors)
  - [Include Context](#include-context)
- [Best Practices](#best-practices)
  - [Exit Codes](#exit-codes)
  - [Error Message Checklist](#error-message-checklist)
  - [Signal-to-Noise Ratio](#signal-to-noise-ratio)
