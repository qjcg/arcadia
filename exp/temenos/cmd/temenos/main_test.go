package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"temenos/internal/storage"
)

func TestAssetsFileServer(t *testing.T) {
	// Create a temp file to serve
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.css"
	if err := os.WriteFile(tmpFile, []byte("/* test */"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set up file server
	fs := http.FileServer(http.Dir(tmpDir))
	handler := http.StripPrefix("/assets/", fs)

	req := httptest.NewRequest(http.MethodGet, "/assets/test.css", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAppHandleHomeGET(t *testing.T) {
	// Create in-memory database for testing
	db, err := storage.OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	app := &App{db: db}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	app.handleHome(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Temenos") {
		t.Error("response should contain 'Temenos'")
	}
}

func TestAppHandleHomeNotAllowed(t *testing.T) {
	db, err := storage.OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	app := &App{db: db}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	app.handleHome(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestAppServeAPI(t *testing.T) {
	db, err := storage.OpenMemory()
	if err != nil {
		t.Fatalf("failed to open memory database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	app := &App{db: db}

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	app.serveAPI(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "ok") {
		t.Error("response should contain 'ok'")
	}
}
