package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"time"

	"temenos/internal/domain"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

//go:embed migrations/*.sql
var migrations embed.FS

type DB struct {
	*Queries
	db *sql.DB
}

func Open(path string) (*DB, error) {
	if path == "" {
		path = "temenos.db"
	}
	database, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Performance PRAGMAs
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000", // 64MB cache
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := database.Exec(p); err != nil {
			return nil, fmt.Errorf("failed to set pragma %s: %w", p, err)
		}
	}

	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		Queries: New(database),
		db:      database,
	}, nil
}

func OpenMemory() (*DB, error) {
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open memory database: %w", err)
	}

	return &DB{
		Queries: New(database),
		db:      database,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Migrate() error {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort migration files by name to ensure consistent order
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		migration, err := migrations.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		_, err = d.db.Exec(string(migration))
		if err != nil {
			return fmt.Errorf("failed to run migration %s: %w", name, err)
		}
	}

	return nil
}

// Domain conversion functions
func (d *DB) CreateModule(ctx context.Context, mod domain.Module) error {
	return d.Queries.CreateModule(ctx, CreateModuleParams{
		ID:          mod.ID,
		Title:       mod.Title,
		Description: nullString(mod.Description),
		Category:    nullString(mod.Category),
		CreatedAt:   mod.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   mod.UpdatedAt.Format(time.RFC3339),
	})
}

func (d *DB) GetModule(ctx context.Context, id string) (*domain.Module, error) {
	row, err := d.Queries.GetModule(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Module{
		ID:          row.ID,
		Title:       row.Title,
		Description: nullStringToString(row.Description),
		Category:    nullStringToString(row.Category),
		CreatedAt:   parseTime(row.CreatedAt),
		UpdatedAt:   parseTime(row.UpdatedAt),
	}, nil
}

func (d *DB) ListModules(ctx context.Context) ([]domain.Module, error) {
	rows, err := d.Queries.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	modules := make([]domain.Module, len(rows))
	for i, row := range rows {
		modules[i] = domain.Module{
			ID:          row.ID,
			Title:       row.Title,
			Description: nullStringToString(row.Description),
			Category:    nullStringToString(row.Category),
			CreatedAt:   parseTime(row.CreatedAt),
			UpdatedAt:   parseTime(row.UpdatedAt),
		}
	}
	return modules, nil
}

func (d *DB) CreateCard(ctx context.Context, card domain.Card) error {
	return d.Queries.CreateCard(ctx, CreateCardParams{
		ID:        card.ID,
		Type:      string(card.Type),
		ModuleID:  card.ModuleID,
		Front:     card.Front,
		Back:      card.Back,
		Metadata:  nullString(""),
		CreatedAt: card.CreatedAt.Format(time.RFC3339),
		UpdatedAt: card.UpdatedAt.Format(time.RFC3339),
	})
}

func (d *DB) GetCard(ctx context.Context, id string) (*domain.Card, error) {
	row, err := d.Queries.GetCard(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.Card{
		ID:       row.ID,
		Type:     domain.CardType(row.Type),
		ModuleID: row.ModuleID,
		Front:    row.Front,
		Back:     row.Back,
	}, nil
}

func (d *DB) ListCardsByModule(ctx context.Context, moduleID string) ([]domain.Card, error) {
	rows, err := d.Queries.ListCardsByModule(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	cards := make([]domain.Card, len(rows))
	for i, row := range rows {
		cards[i] = domain.Card{
			ID:       row.ID,
			Type:     domain.CardType(row.Type),
			ModuleID: row.ModuleID,
			Front:    row.Front,
			Back:     row.Back,
		}
	}
	return cards, nil
}

func (d *DB) ListCardsForReview(ctx context.Context, moduleID string, limit int) ([]domain.Card, error) {
	rows, err := d.Queries.ListCardsForReview(ctx, ListCardsForReviewParams{
		ModuleID:     moduleID,
		NextReviewAt: time.Now().Format(time.RFC3339),
		Limit:        int64(limit),
	})
	if err != nil {
		return nil, err
	}
	cards := make([]domain.Card, len(rows))
	for i, row := range rows {
		cards[i] = domain.Card{
			ID:       row.ID,
			Type:     domain.CardType(row.Type),
			ModuleID: row.ModuleID,
			Front:    row.Front,
			Back:     row.Back,
		}
	}
	return cards, nil
}

func (d *DB) CountCardsByModule(ctx context.Context, moduleID string) (int, error) {
	count, err := d.Queries.CountCardsByModule(ctx, moduleID)
	return int(count), err
}

func (d *DB) CountReviews(ctx context.Context) (int, error) {
	count, err := d.Queries.CountReviews(ctx)
	return int(count), err
}

func (d *DB) CountReviewsByModule(ctx context.Context, moduleID string) (int, error) {
	count, err := d.Queries.CountReviewsByModule(ctx, moduleID)
	return int(count), err
}

func (d *DB) CountCardsDueForReview(ctx context.Context) (int, error) {
	count, err := d.Queries.CountCardsDueForReview(ctx, time.Now().Format(time.RFC3339))
	return int(count), err
}

func (d *DB) CountCardsDueForReviewByModule(ctx context.Context, moduleID string) (int, error) {
	count, err := d.Queries.CountCardsDueForReviewByModule(ctx, CountCardsDueForReviewByModuleParams{
		ModuleID:     moduleID,
		NextReviewAt: time.Now().Format(time.RFC3339),
	})
	return int(count), err
}

type ModuleStats struct {
	TotalCards    int
	ReviewedCards int
	AvgQuality    float64
}

func (d *DB) GetModuleStats(ctx context.Context, moduleID string) (*ModuleStats, error) {
	row, err := d.Queries.GetModuleStats(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	avgQ, _ := row.AvgQuality.(float64)
	return &ModuleStats{
		TotalCards:    int(row.TotalCards),
		ReviewedCards: int(row.ReviewedCards),
		AvgQuality:    avgQ,
	}, nil
}

func (d *DB) GetAllModuleStats(ctx context.Context) (map[string]*ModuleStats, error) {
	rows, err := d.Queries.GetAllModuleStats(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*ModuleStats, len(rows))
	for _, row := range rows {
		avgQ, _ := row.AvgQuality.(float64)
		result[row.ModuleID] = &ModuleStats{
			TotalCards:    int(row.TotalCards),
			ReviewedCards: int(row.ReviewedCards),
			AvgQuality:    avgQ,
		}
	}
	return result, nil
}

func (d *DB) CreateReview(ctx context.Context, review domain.Review) error {
	return d.Queries.CreateReview(ctx, CreateReviewParams{
		ID:           review.ID,
		CardID:       review.CardID,
		SessionID:    nullString(review.SessionID),
		Quality:      int64(review.Quality),
		IntervalDays: int64(review.Interval),
		EaseFactor:   review.EaseFactor,
		ReviewCount:  int64(review.ReviewCount),
		NextReviewAt: review.NextReviewAt.Format(time.RFC3339),
		ReviewedAt:   review.ReviewedAt.Format(time.RFC3339),
	})
}

func (d *DB) GetReviewForCard(ctx context.Context, cardID string) (*domain.Review, error) {
	row, err := d.Queries.GetReviewForCard(ctx, cardID)
	if err != nil {
		return nil, err
	}
	return &domain.Review{
		ID:           row.ID,
		CardID:       row.CardID,
		SessionID:    nullStringToString(row.SessionID),
		Quality:      int(row.Quality),
		Interval:     int(row.IntervalDays),
		EaseFactor:   row.EaseFactor,
		ReviewCount:  int(row.ReviewCount),
		NextReviewAt: parseTime(row.NextReviewAt),
		ReviewedAt:   parseTime(row.ReviewedAt),
	}, nil
}

func (d *DB) UpdateReview(ctx context.Context, review domain.Review) error {
	return d.Queries.UpdateReview(ctx, UpdateReviewParams{
		Quality:      int64(review.Quality),
		IntervalDays: int64(review.Interval),
		EaseFactor:   review.EaseFactor,
		ReviewCount:  int64(review.ReviewCount),
		NextReviewAt: review.NextReviewAt.Format(time.RFC3339),
		ReviewedAt:   review.ReviewedAt.Format(time.RFC3339),
		ID:           review.ID,
	})
}

func (d *DB) CreateSession(ctx context.Context, session domain.Session) error {
	var endedAt *string
	if session.EndedAt != nil && !session.EndedAt.IsZero() {
		s := session.EndedAt.Format(time.RFC3339)
		endedAt = &s
	}
	return d.Queries.CreateSession(ctx, CreateSessionParams{
		ID:        session.ID,
		ModuleID:  session.ModuleID,
		Mode:      string(session.Mode),
		StartedAt: session.StartedAt.Format(time.RFC3339),
		EndedAt:   nullStringPtr(endedAt),
	})
}

func (d *DB) GetSession(ctx context.Context, id string) (*domain.Session, error) {
	row, err := d.Queries.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	startedAt := parseTime(row.StartedAt)
	return &domain.Session{
		ID:        row.ID,
		ModuleID:  row.ModuleID,
		Mode:      domain.SessionMode(row.Mode),
		StartedAt: startedAt,
		EndedAt:   nilOrParseTime(row.EndedAt),
	}, nil
}

// Helper functions
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullStringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullStringToStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		log.Printf("failed to parse time %q: %v", s, err)
		return time.Time{}
	}
	return t
}

func nilOrParseTime(ns sql.NullString) *time.Time {
	if !ns.Valid {
		return nil
	}
	t := parseTime(ns.String)
	return &t
}

// CalculateStreak calculates the current streak (consecutive days with reviews)
func (d *DB) CalculateStreak(ctx context.Context) (int, error) {
	dates, err := d.Queries.GetReviewDates(ctx)
	if err != nil {
		return 0, err
	}
	if len(dates) == 0 {
		return 0, nil
	}

	// Group reviews by date
	reviewDates := make(map[string]bool)
	for _, dateStr := range dates {
		dt, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			continue
		}
		reviewDates[dt.Format("2006-01-02")] = true
	}

	// Count consecutive days from today going backwards
	streak := 0
	current := time.Now()
	for {
		dateStr := current.Format("2006-01-02")
		if reviewDates[dateStr] {
			streak++
			current = current.AddDate(0, 0, -1)
		} else {
			// Allow for today not being reviewed yet
			if streak == 0 {
				current = current.AddDate(0, 0, -1)
				dateStr = current.Format("2006-01-02")
				if reviewDates[dateStr] {
					streak++
					current = current.AddDate(0, 0, -1)
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	return streak, nil
}
