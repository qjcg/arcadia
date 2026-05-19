# Output Design

Design CLI output for both humans and machines.

## Principles

### Human-Readable by Default
- Output should be understandable at a glance
- Use formatting (whitespace, indentation) to aid scanning
- Avoid raw technical jargon in user-facing output

### Machine-Readable on Request
- Provide `--json` or `--format json` for scripting
- Ensure JSON output is valid and consistent
- Include all relevant data in JSON mode

## Best Practices

### Color and Formatting
```
# Good: Color is optional and disabled when not a TTY
if stdout.isTTY() {
    fmt.Fprintf彩色输出()
}

# Good: --no-color disables all formatting
if flagNoColor {
    color.NoColor = true
}
```

### Progress for Long Operations
```
# Good: Show progress for operations > 2 seconds
cmd.Flags().Bool("progress", isTTY, "Show progress bar")

# Patterns:
# - spinners for indeterminate duration
# - progress bars for known duration
# - step indicators for discrete steps: [1/5] [2/5]
```

### Verbosity Levels
| Flag | Effect |
|------|--------|
| `-q`, `--quiet` | Only errors and final result |
| (default) | Normal output |
| `-v`, `--verbose` | Additional context |
| `-vv` or `--debug` | Full debug info |

### Structured Output
```bash
# Provide machine-readable options
--output json     # JSON lines
--output yaml     # YAML
--output format   # Custom format string
```

## Table of Contents

- [Principles](#principles)
- [Best Practices](#best-practices)
  - [Color and Formatting](#color-and-formatting)
  - [Progress for Long Operations](#progress-for-long-operations)
  - [Verbosity Levels](#verbosity-levels)
  - [Structured Output](#structured-output)
