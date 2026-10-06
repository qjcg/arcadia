# Compatibility & Stability

Maintain compatibility as your CLI evolves.

## The Compatibility Contract

Users build scripts and workflows around your CLI. Breaking changes cost them time.

### Additive is Safe
Adding new flags, commands, or output fields is safe. Users who don't use them are unaffected.

### Breaking Changes Require Warning

```bash
# Good: Deprecation warning
$ myapp old-command
Warning: 'old-command' is deprecated and will be removed in v3.0.
Use 'new-command' instead.
Run 'myapp migrate --help' for migration instructions.
```

### Breaking Change Checklist
- [ ] Announce in release notes
- [ ] Print deprecation warning for 2+ minor versions
- [ ] Provide migration path
- [ ] Consider aliases for old patterns

## Version Handling

### `--version` Output
```
myapp version 2.1.0 (build: abc123)
```

### Semantic Versioning
- **Major** (2.0.0): Breaking changes
- **Minor** (2.1.0): New features, backward compatible
- **Patch** (2.1.1): Bug fixes

### Time Bombs
Avoid creating time bombs — features that work now but fail at a future date:

```go
// Bad: Will break in 2025
if time.Now().Year() > 2025 {
    return errors.New("license expired")
}

// Good: Version-based deprecation
if viper.GetBool("legacy_mode") {
    return warnLegacyMode()
}
```

## Cross-Platform Considerations

### File Paths
```go
// Bad: Unix-specific
path := "/tmp/file"

// Good: Use os.TempDir()
path := filepath.Join(os.TempDir(), "file")

// Good: Handle both separators
path := filepath.Join(homedir, ".myapp", "config")
```

### Shell Compatibility
- Test with bash, zsh, fish
- Consider POSIX compatibility for scripts
- Handle `set -e` behavior correctly

## Table of Contents

- [The Compatibility Contract](#the-compatibility-contract)
  - [Additive is Safe](#additive-is-safe)
  - [Breaking Changes Require Warning](#breaking-changes-require-warning)
  - [Breaking Change Checklist](#breaking-change-checklist)
- [Version Handling](#version-handling)
  - [--version Output](#--version-output)
  - [Semantic Versioning](#semantic-versioning)
  - [Time Bombs](#time-bombs)
- [Cross-Platform Considerations](#cross-platform-considerations)
  - [File Paths](#file-paths)
  - [Shell Compatibility](#shell-compatibility)
