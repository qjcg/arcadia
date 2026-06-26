package cli

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/fsnotify/fsnotify"
	"github.com/qjcg/arcadia/exp/pavona/pkg/site"
	"github.com/spf13/cobra"
)

type ServeParams struct {
	Dir  string `short:"d" descr:"Project directory" default:"." optional:"true"`
	Port int    `short:"p" descr:"Port to serve on" default:"8080" optional:"true"`
}

func ServeCmd() boa.CmdT[ServeParams] {
	return boa.CmdT[ServeParams]{
		Use:   "serve",
		Short: "Start the dev server with live reload",
		Long:  "Start a development server that rebuilds the site on changes.",
		RunFunc: func(p *ServeParams, cmd *cobra.Command, args []string) {
			contentDir := filepath.Join(p.Dir, "content")
			outputDir := filepath.Join(p.Dir, "dist")

			if err := site.Build(contentDir, outputDir); err != nil {
				fmt.Fprintf(os.Stderr, "Error building site: %v\n", err)
				os.Exit(1)
			}

			// Start file watcher for live rebuilds
			go watchContent(contentDir, outputDir)

			fs := http.FileServer(http.Dir(outputDir))
			mux := http.NewServeMux()
			mux.Handle("/", fs)

			addr := fmt.Sprintf(":%d", p.Port)
			slog.Info("dev server starting", "addr", addr, "dir", outputDir)
			if err := http.ListenAndServe(addr, mux); err != nil {
				slog.Error("server failed", "error", err)
				os.Exit(1)
			}
		},
	}
}

func watchContent(contentDir, outputDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Warn("file watching unavailable", "error", err)
		return
	}
	defer watcher.Close()

	// Watch the content directory and all subdirectories
	if err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return err
		}
		return watcher.Add(path)
	}); err != nil {
		slog.Warn("failed to watch content directory", "error", err)
		return
	}

	extensions := map[string]bool{".md": true, ".org": true}
	var debounce *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) == 0 {
				continue
			}
			if !extensions[strings.ToLower(filepath.Ext(event.Name))] {
				continue
			}
			// Debounce: reset timer on each event, rebuild 200ms after last one
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(200*time.Millisecond, func() {
				slog.Info("content changed, rebuilding...", "file", event.Name)
				if err := site.Build(contentDir, outputDir); err != nil {
					slog.Error("rebuild failed", "error", err)
				} else {
					slog.Info("rebuild complete")
				}
			})
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Warn("file watcher error", "error", err)
		}
	}
}
