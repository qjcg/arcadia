package cli

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/bolocera/pavona/internal/site"
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
