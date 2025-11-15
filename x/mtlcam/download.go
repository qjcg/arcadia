package main

import (
	"io"
	"net/http"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// GET and return contents from URL.
func download(url string) []byte {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		log.Fatal("Couldn't GET file.")
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Couldn't read response body.")
		return nil
	}

	return contents
}

// Write slice of bytes to disk.
func saveFile(data []byte, filename, dir string) {
	err := os.WriteFile(path.Join(dir, filename), data, 0o644)
	if err != nil {
		log.Fatal("Couldn't create file -- ", err)
	}
}
