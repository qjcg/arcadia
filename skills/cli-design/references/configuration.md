# Configuration

Handle configuration with clear precedence and sensible defaults.

## Precedence Order

Users expect this priority (highest to lowest):

1. **Flags** — Explicit command-line arguments
2. **Environment variables** — Context-specific overrides
3. **Project config** — `.env`, `config.yaml` in project root
4. **User config** — `~/.config/app/` (XDG standard)
5. **System config** — `/etc/app/` (if applicable)

## Environment Variables

### Naming
```
# Good: Uppercase with prefix
MYAPP_DEBUG=true
MYAPP_DATABASE_URL=postgres://...

# Use viper's automatic env binding
viper.SetEnvPrefix("MYAPP")
viper.AutomaticEnv()
```

### Documentation
Document all env vars in help text:
```
Environment variables:
  MYAPP_DEBUG      Enable debug logging (default: false)
  MYAPP_LOG_LEVEL  Log level: debug, info, warn, error (default: info)
  MYAPP_TIMEOUT    Request timeout in seconds (default: 30)
```

## Config Files

### XDG Base Directory
```go
import "github.com/mitchellh/go-homedir"

func getConfigPath() string {
    // ~/.config/app/ on Unix
    // %APPDATA%/app on Windows
    configDir := os.Getenv("XDG_CONFIG_HOME")
    if configDir == "" {
        homedir, _ := homedir.Dir()
        configDir = filepath.Join(homedir, ".config")
    }
    return filepath.Join(configDir, "myapp")
}
```

### Config File Formats
- YAML for complex config
- JSON for programmatic config
- TOML for simpler config with good human readability
- `.env` files for local development

## Sensible Defaults

Default values should "just work" for common cases:

```go
// Bad: Requires user to always specify
cmd.Flags().String("database", "", "Database URL (REQUIRED)")

// Good: Sensible default
cmd.Flags().String("database", "sqlite://data.db", "Database URL")

// Good: Default to localhost for dev
cmd.Flags().String("host", "localhost", "Server host")
cmd.Flags().Int("port", 8080, "Server port")
```

## Table of Contents

- [Precedence Order](#precedence-order)
- [Environment Variables](#environment-variables)
  - [Naming](#naming)
  - [Documentation](#documentation)
- [Config Files](#config-files)
  - [XDG Base Directory](#xdg-base-directory)
  - [Config File Formats](#config-file-formats)
- [Sensible Defaults](#sensible-defaults)
