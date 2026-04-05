package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"temenos/internal/storage"
)

// Config holds web server configuration
type Config struct {
	Port       string
	DBPath     string
	AssetsPath string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	return &Config{
		Port:       getEnvOrDefault("PORT", "8080"),
		DBPath:     getEnvOrDefault("DB_PATH", "temenos.db"),
		AssetsPath: getEnvOrDefault("ASSETS_PATH", "assets"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Server represents the HTTP server
type Server struct {
	httpServer *http.Server
	db         *storage.DB
	config     *Config
}

// New creates a new web server
func New(cfg *Config, db *storage.DB) *Server {
	return &Server{
		db:     db,
		config: cfg,
	}
}

// Router returns the HTTP handler with pattern-based routing
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Static asset serving
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(s.config.AssetsPath))))

	return mux
}

// Start starts the web server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Starting temenos on :%s", s.config.Port)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// StartWithGracefulShutdown starts the server with graceful shutdown handling
func (s *Server) StartWithGracefulShutdown() error {
	s.httpServer = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.Router(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("Starting temenos on :%s", s.config.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server stopped")
	return nil
}

// Handler returns the http.Handler for the server
func (s *Server) Handler() http.Handler {
	return s.Router()
}

// DB returns the database connection
func (s *Server) DB() *storage.DB {
	return s.db
}

// Config returns the server config
func (s *Server) Config() *Config {
	return s.config
}
