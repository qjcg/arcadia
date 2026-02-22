package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

// Serve starts a web server on the given port to serve the embedded WASM application.
func Serve(port int) error {
	sub, err := fs.Sub(assets, "assets")
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(sub)))

	fmt.Printf("Serving fractalis at http://localhost:%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
