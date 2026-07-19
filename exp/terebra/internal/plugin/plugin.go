package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plugin"
	"sort"
	"strings"
)

// PluginHandler is the interface that plugins must implement.
type PluginHandler interface {
	// Name returns the command name for this plugin.
	Name() string
	// Execute runs the plugin command.
	Execute(args []string, stdin io.Reader, stdout, stderr io.Writer) int
}

// Registry manages loaded plugins.
type Registry struct {
	plugins map[string]PluginHandler
	paths   []string
}

// New creates a new plugin registry.
func New() *Registry {
	r := &Registry{
		plugins: make(map[string]PluginHandler),
	}
	// Default plugin search paths
	if home := os.Getenv("HOME"); home != "" {
		r.paths = append(r.paths, filepath.Join(home, ".terebra", "plugins"))
	}
	r.paths = append(r.paths, "/usr/lib/terebra/plugins")
	r.paths = append(r.paths, "/usr/local/lib/terebra/plugins")
	if env := os.Getenv("TEREBRA_PLUGIN_PATH"); env != "" {
		r.paths = append(r.paths, filepath.SplitList(env)...)
	}
	return r
}

// LoadAll loads all plugins from the search paths.
func (r *Registry) LoadAll() {
	for _, dir := range r.paths {
		r.loadDir(dir)
	}
}

// Load loads a single plugin .so file.
func (r *Registry) Load(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("cannot load plugin %q: %v", path, err)
	}

	sym, err := p.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("plugin %q: missing Plugin symbol: %v", path, err)
	}

	handler, ok := sym.(PluginHandler)
	if !ok {
		return fmt.Errorf("plugin %q: Plugin symbol does not implement PluginHandler", path)
	}

	name := handler.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %q: command %q already registered", path, name)
	}

	r.plugins[name] = handler
	return nil
}

// loadDir loads all .so files from a directory.
func (r *Registry) loadDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".so") {
			continue
		}
		r.Load(filepath.Join(dir, e.Name()))
	}
}

// Lookup returns a plugin handler by name.
func (r *Registry) Lookup(name string) (PluginHandler, bool) {
	h, ok := r.plugins[name]
	return h, ok
}

// List returns all registered plugin names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.plugins))
	for n := range r.plugins {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// PluginBuiltin returns a builtin-compatible handler that dispatches to plugins.
func PluginBuiltin(registry *Registry) func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return func(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
		if len(args) == 0 {
			fmt.Fprintln(stderr, "plugin: expected subcommand (load, list)")
			fmt.Fprintln(stderr, "  plugin load <path>  Load a plugin .so file")
			fmt.Fprintln(stderr, "  plugin list         List loaded plugins")
			return 1
		}

		switch args[0] {
		case "load":
			if len(args) < 2 {
				fmt.Fprintln(stderr, "plugin load: expected path")
				return 1
			}
			if err := registry.Load(args[1]); err != nil {
				fmt.Fprintf(stderr, "plugin load: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "loaded plugin: %s\n", args[1])
			return 0

		case "list":
			plugins := registry.List()
			if len(plugins) == 0 {
				fmt.Fprintln(stdout, "no plugins loaded")
			}
			for _, name := range plugins {
				fmt.Fprintf(stdout, "  %s\n", name)
			}
			return 0

		default:
			// Try to execute a plugin command
			if handler, ok := registry.Lookup(args[0]); ok {
				return handler.Execute(args[1:], stdin, stdout, stderr)
			}
			fmt.Fprintf(stderr, "plugin: unknown subcommand: %s\n", args[0])
			return 1
		}
	}
}
