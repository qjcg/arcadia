# Flag Design

Design flags that are intuitive and follow conventions.

## Principles

### Flags Over Positional Arguments
Positional arguments are harder to remember and less flexible.

```
# Bad: What does '60' mean?
myapp upload file.txt 60

# Good: Self-documenting
myapp upload file.txt --timeout 60
```

### When to Use Positional Arguments
- Required values that are central to the command
- Paths or file references
- Things that form a natural "sentence" with the command

### Boolean Flags
```go
// Bad: What does --force=false mean?
cmd.Flags().Bool("force", false, "Force operation")

// Good: --force is present = true, absent = false
// This is the Go bool flag default behavior, just ensure documentation is clear
```

## Naming Conventions

### Standard Flags (Follow These)
| Flag | Meaning |
|------|---------|
| `-h`, `--help` | Show help |
| `-v`, `--version` | Show version |
| `-c`, `--config` | Config file path |
| `-o`, `--output` | Output file |
| `-f`, `--force` | Overwrite without prompting |
| `-q`, `--quiet` | Suppress output |
| `-V`, `--verbose` | Increase verbosity |

### Long Flag Names
- Use kebab-case: `--max-retries`, `--output-file`
- Make them readable and descriptive
- Avoid abbreviations unless very common

### Short Flags
| Short | Common Use |
|-------|------------|
| `-c` | Config |
| `-f` | File/Force |
| `-h` | Help |
| `-o` | Output |
| `-v` | Verbose/Version |
| `-q` | Quiet |

## Flag Types

### String Flags
```go
cmd.Flags().String("format", "text", "Output format (text, json, yaml)")
```

### Integer Flags
```go
cmd.Flags().Int("port", 8080, "Port to listen on")
cmd.Flags().Int("max-retries", 3, "Maximum retry attempts")
```

### Boolean Flags
```go
cmd.Flags().Bool("force", false, "Skip confirmation prompts")
cmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
```

### Enum-like Flags (String with restricted values)
```go
// Validate in PreRunE
cmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error)")
```

## stdin/stdout Convention

Use `-` to represent stdin/stdout:

```bash
# Read from stdin
cat file.txt | myapp process -

# Write to stdout
myapp export --format json - > output.json

# Both
myapp transform - -
```

## Flag Documentation

```go
cmd.Flags().String("output", "result.txt",
    "Output file path.\n"+
    "Use '-' for stdout.\n"+
    "Defaults to 'result.txt' in current directory.")
```

## Table of Contents

- [Principles](#principles)
  - [Flags Over Positional Arguments](#flags-over-positional-arguments)
  - [When to Use Positional Arguments](#when-to-use-positional-arguments)
  - [Boolean Flags](#boolean-flags)
- [Naming Conventions](#naming-conventions)
  - [Standard Flags (Follow These)](#standard-flags-follow-these)
  - [Long Flag Names](#long-flag-names)
  - [Short Flags](#short-flags)
- [Flag Types](#flag-types)
  - [String Flags](#string-flags)
  - [Integer Flags](#integer-flags)
  - [Boolean Flags](#boolean-flags)
  - [Enum-like Flags (String with restricted values)](#enum-like-flags-string-with-restricted-values)
- [stdin/stdout Convention](#stdinstdout-convention)
- [Flag Documentation](#flag-documentation)
