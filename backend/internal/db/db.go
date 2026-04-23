package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store groups together the database operations the application can perform.
// This gives handlers a small, explicit interface instead of coupling them
// directly to raw SQL and the connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// TournamentConfig mirrors the fields the frontend will need from the
// tournament_config table. JSONB columns are decoded into generic maps for now
// so the API can return flexible year-to-year configuration data.
type TournamentConfig struct {
	Year              int            `json:"year"`
	EntryDeadline     *time.Time     `json:"entry_deadline"`
	StartDate         *time.Time     `json:"start_date"`
	EndDate           *time.Time     `json:"end_date"`
	Groups            map[string]any `json:"groups"`
	MuttMultiplier    string         `json:"mutt_multiplier"`
	OldMuttMultiplier string         `json:"old_mutt_multiplier"`
	PoolPayouts       map[string]any `json:"pool_payouts"`
	FRLWinner         *string        `json:"frl_winner"`
	FRLPayout         int            `json:"frl_payout"`
	Active            bool           `json:"active"`
}

var ErrNotFound = errors.New("not found")

// NewPool creates the shared Postgres connection pool for the application.
// Using a pool from the start matches how the app will behave in production
// and gives us one place to tune database settings later.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	// Ping early so startup fails fast if the app cannot reach Postgres.
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return pool, nil
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// GetConfig returns the tournament configuration row for a given year.
// This is the first typed query in the project and acts as the pattern future
// DB functions should follow: accept context, return a domain-shaped struct,
// and hide SQL details from the caller.
func (s *Store) GetConfig(ctx context.Context, year int) (TournamentConfig, error) {
	const query = `
		SELECT
			year,
			entry_deadline,
			start_date,
			end_date,
			groups,
			mutt_multiplier::text,
			old_mutt_multiplier::text,
			pool_payouts,
			frl_winner,
			frl_payout,
			active
		FROM tournament_config
		WHERE year = $1
	`

	var cfg TournamentConfig
	var groupsRaw []byte
	var payoutsRaw []byte

	err := s.pool.QueryRow(ctx, query, year).Scan(
		&cfg.Year,
		&cfg.EntryDeadline,
		&cfg.StartDate,
		&cfg.EndDate,
		&groupsRaw,
		&cfg.MuttMultiplier,
		&cfg.OldMuttMultiplier,
		&payoutsRaw,
		&cfg.FRLWinner,
		&cfg.FRLPayout,
		&cfg.Active,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TournamentConfig{}, ErrNotFound
		}

		return TournamentConfig{}, fmt.Errorf("get tournament config for year %d: %w", year, err)
	}

	if len(groupsRaw) > 0 {
		if err := json.Unmarshal(groupsRaw, &cfg.Groups); err != nil {
			return TournamentConfig{}, fmt.Errorf("decode groups json: %w", err)
		}
	}

	if len(payoutsRaw) > 0 {
		if err := json.Unmarshal(payoutsRaw, &cfg.PoolPayouts); err != nil {
			return TournamentConfig{}, fmt.Errorf("decode payouts json: %w", err)
		}
	}

	return cfg, nil
}
