package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"temenos/cmd/temenos/templates"
	"temenos/internal/domain"
	"temenos/internal/storage"
)

// Config holds application configuration
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

// App holds dependencies for the HTTP handlers
type App struct {
	db     *storage.DB
	config *Config
}

func main() {
	cfg := LoadConfig()

	db, err := storage.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	// Seed is opt-in via SEED=1 environment variable
	if os.Getenv("SEED") == "1" {
		if err := seed(db); err != nil {
			log.Fatalf("failed to seed: %v", err)
		}
	}

	app := &App{db: db, config: cfg}

	// Use ServeMux for routing
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleHome)
	mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(cfg.AssetsPath))))

	// Create server with timeouts
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("Starting temenos on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func (app *App) handleHome(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Only handle GET for now
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Pattern matching for routes
	switch {
	case path == "/", path == "/modules":
		app.serveModules(w, r)
	case path == "/progress":
		app.serveProgress(w, r)
	default:
		switch {
		case len(path) > 8 && path[:8] == "/modules/":
			app.serveModule(w, r, path)
		case len(path) > 6 && path[:6] == "/study/":
			app.serveStudy(w, r, path)
		case len(path) > 4 && path[:4] == "/api/":
			app.serveAPI(w, r)
		default:
			http.NotFound(w, r)
		}
	}
}

func (app *App) serveModules(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	modules, err := app.db.ListModules(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to domain modules
	domainModules := make([]domain.Module, len(modules))
	for i, m := range modules {
		domainModules[i] = domain.Module{
			ID:          m.ID,
			Title:       m.Title,
			Description: m.Description,
		}
	}

	templates.Modules(domainModules).Render(r.Context(), w)
}

func (app *App) serveModule(w http.ResponseWriter, r *http.Request, path string) {
	moduleID := path[8:] // remove "/modules/" prefix
	if moduleID == "" {
		http.NotFound(w, r)
		return
	}

	ctx := context.Background()

	mod, err := app.db.GetModule(ctx, moduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	cards, err := app.db.ListCardsByModule(ctx, moduleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to domain types
	domainMod := domain.Module{
		ID:          mod.ID,
		Title:       mod.Title,
		Description: mod.Description,
	}
	domainCards := make([]domain.Card, len(cards))
	for i, c := range cards {
		domainCards[i] = domain.Card{
			ID:    c.ID,
			Front: c.Front,
			Back:  c.Back,
		}
	}

	templates.ModuleDetail(domainMod, domainCards, moduleID).Render(r.Context(), w)
}

func (app *App) serveProgress(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	totalReviews, err := app.db.CountReviews(ctx)
	if err != nil {
		log.Printf("error counting reviews: %v", err)
		totalReviews = 0
	}
	cardsDue, err := app.db.CountCardsDueForReview(ctx)
	if err != nil {
		log.Printf("error counting cards due: %v", err)
		cardsDue = 0
	}

	streak, err := app.db.CalculateStreak(ctx)
	if err != nil {
		log.Printf("error calculating streak: %v", err)
		streak = 0
	}

	modules, err := app.db.ListModules(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allStats, err := app.db.GetAllModuleStats(ctx)
	if err != nil {
		log.Printf("error getting module stats: %v", err)
		allStats = nil
	}

	templates.Progress(totalReviews, cardsDue, streak, modules, allStats).Render(r.Context(), w)
}

func (app *App) serveAPI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case path == "/api/review":
		app.handleReviewAPI(w, r)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}
