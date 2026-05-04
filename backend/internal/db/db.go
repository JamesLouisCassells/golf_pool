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

type User struct {
	ClerkID     string    `json:"clerk_id"`
	Email       string    `json:"email"`
	DisplayName *string   `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type Entry struct {
	ID          string         `json:"id"`
	Year        int            `json:"year"`
	ClerkID     *string        `json:"clerk_id"`
	DisplayName string         `json:"display_name"`
	Picks       map[string]any `json:"picks"`
	InOvers     bool           `json:"in_overs"`
	Locked      bool           `json:"locked"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
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

type UpdateTournamentConfigParams struct {
	Year              int
	EntryDeadline     *time.Time
	StartDate         *time.Time
	EndDate           *time.Time
	Groups            map[string]any
	MuttMultiplier    string
	OldMuttMultiplier string
	PoolPayouts       map[string]any
	FRLWinner         *string
	FRLPayout         int
	Active            bool
}

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

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

// GetUser returns the local user record tied to a Clerk user ID.
// Keeping this in the store lets handlers ask for a user in domain terms
// instead of repeating SQL every time we need identity data.
func (s *Store) GetUser(ctx context.Context, clerkID string) (User, error) {
	const query = `
		SELECT clerk_id, email, display_name, created_at
		FROM users
		WHERE clerk_id = $1
	`

	var user User

	err := s.pool.QueryRow(ctx, query, clerkID).Scan(
		&user.ClerkID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}

		return User{}, fmt.Errorf("get user %s: %w", clerkID, err)
	}

	return user, nil
}

// UpsertUser creates or refreshes the local user record from the identity
// provider claims. This gives us a single place to keep user basics in sync
// whenever an authenticated request reaches the API.
func (s *Store) UpsertUser(ctx context.Context, clerkID, email string, displayName *string) (User, error) {
	const query = `
		INSERT INTO users (clerk_id, email, display_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (clerk_id)
		DO UPDATE SET
			email = EXCLUDED.email,
			display_name = EXCLUDED.display_name
		RETURNING clerk_id, email, display_name, created_at
	`

	var user User

	err := s.pool.QueryRow(ctx, query, clerkID, email, displayName).Scan(
		&user.ClerkID,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("upsert user %s: %w", clerkID, err)
	}

	return user, nil
}

// GetMyEntry returns the authenticated user's entry for the currently active
// tournament year. Joining through tournament_config keeps "current year"
// logic in one place instead of pushing that rule into handlers.
func (s *Store) GetMyEntry(ctx context.Context, clerkID string) (Entry, error) {
	const query = `
		SELECT
			e.id::text,
			e.year,
			e.clerk_id,
			e.display_name,
			e.picks,
			e.in_overs,
			e.locked,
			e.created_at,
			e.updated_at
		FROM entries e
		INNER JOIN tournament_config tc
			ON tc.year = e.year
		WHERE e.clerk_id = $1
			AND tc.active = true
		ORDER BY e.updated_at DESC
		LIMIT 1
	`

	entry, err := s.scanEntry(ctx, query, clerkID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Entry{}, ErrNotFound
		}
		return Entry{}, fmt.Errorf("get active entry for user %s: %w", clerkID, err)
	}

	return entry, nil
}

// ListEntriesForActiveYear returns every entry for the currently active
// tournament year. This supports the public entries view once the tournament
// has started.
func (s *Store) ListEntriesForActiveYear(ctx context.Context) ([]Entry, error) {
	const query = `
		SELECT
			e.id::text,
			e.year,
			e.clerk_id,
			e.display_name,
			e.picks,
			e.in_overs,
			e.locked,
			e.created_at,
			e.updated_at
		FROM entries e
		INNER JOIN tournament_config tc
			ON tc.year = e.year
		WHERE tc.active = true
		ORDER BY e.display_name ASC, e.created_at ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list entries for active year: %w", err)
	}
	defer rows.Close()

	entries := []Entry{}
	for rows.Next() {
		entry, err := scanEntryRow(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active entries: %w", err)
	}

	return entries, nil
}

// GetActiveConfig returns the config row for the tournament year currently
// marked active. This is the config entry routes should use when operating on
// the live pool instead of requiring clients to supply a year separately.
func (s *Store) GetActiveConfig(ctx context.Context) (TournamentConfig, error) {
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
		WHERE active = true
		ORDER BY year DESC
		LIMIT 1
	`

	return s.getConfigByQuery(ctx, query)
}

type CreateEntryParams struct {
	Year        int
	ClerkID     string
	DisplayName string
	Picks       map[string]any
	InOvers     bool
}

type UpdateEntryParams struct {
	ID          string
	DisplayName string
	Picks       map[string]any
	InOvers     bool
}

// CreateEntry inserts a new entry tied to a user and year. A small duplicate
// check here prevents race conditions from slipping past the handler's earlier
// existence check.
func (s *Store) CreateEntry(ctx context.Context, params CreateEntryParams) (Entry, error) {
	const duplicateQuery = `
		SELECT 1
		FROM entries
		WHERE year = $1 AND clerk_id = $2
		LIMIT 1
	`

	var duplicate int
	err := s.pool.QueryRow(ctx, duplicateQuery, params.Year, params.ClerkID).Scan(&duplicate)
	if err == nil {
		return Entry{}, ErrConflict
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return Entry{}, fmt.Errorf("check duplicate entry for user %s year %d: %w", params.ClerkID, params.Year, err)
	}

	picksJSON, err := json.Marshal(params.Picks)
	if err != nil {
		return Entry{}, fmt.Errorf("encode entry picks json: %w", err)
	}

	const insertQuery = `
		INSERT INTO entries (year, clerk_id, display_name, picks, in_overs)
		VALUES ($1, $2, $3, $4::jsonb, $5)
		RETURNING
			id::text,
			year,
			clerk_id,
			display_name,
			picks,
			in_overs,
			locked,
			created_at,
			updated_at
	`

	var entry Entry
	var picksRaw []byte
	err = s.pool.QueryRow(
		ctx,
		insertQuery,
		params.Year,
		params.ClerkID,
		params.DisplayName,
		string(picksJSON),
		params.InOvers,
	).Scan(
		&entry.ID,
		&entry.Year,
		&entry.ClerkID,
		&entry.DisplayName,
		&picksRaw,
		&entry.InOvers,
		&entry.Locked,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		return Entry{}, fmt.Errorf("create entry for user %s year %d: %w", params.ClerkID, params.Year, err)
	}

	if err := json.Unmarshal(picksRaw, &entry.Picks); err != nil {
		return Entry{}, fmt.Errorf("decode inserted entry picks json: %w", err)
	}

	return entry, nil
}

// GetEntryByID returns a single entry regardless of user ownership. Handlers
// can layer permission checks on top of this without duplicating fetch logic.
func (s *Store) GetEntryByID(ctx context.Context, id string) (Entry, error) {
	const query = `
		SELECT
			id::text,
			year,
			clerk_id,
			display_name,
			picks,
			in_overs,
			locked,
			created_at,
			updated_at
		FROM entries
		WHERE id = $1
	`

	return s.scanEntry(ctx, query, id)
}

// UpdateEntry updates the editable parts of an existing entry. Ownership and
// deadline checks stay in the handler layer so the permission rules remain
// explicit at the HTTP boundary.
func (s *Store) UpdateEntry(ctx context.Context, params UpdateEntryParams) (Entry, error) {
	picksJSON, err := json.Marshal(params.Picks)
	if err != nil {
		return Entry{}, fmt.Errorf("encode entry picks json: %w", err)
	}

	const query = `
		UPDATE entries
		SET
			display_name = $2,
			picks = $3::jsonb,
			in_overs = $4,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id::text,
			year,
			clerk_id,
			display_name,
			picks,
			in_overs,
			locked,
			created_at,
			updated_at
	`

	entry, err := s.scanEntry(ctx, query, params.ID, params.DisplayName, string(picksJSON), params.InOvers)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Entry{}, ErrNotFound
		}
		return Entry{}, fmt.Errorf("update entry %s: %w", params.ID, err)
	}

	return entry, nil
}

// DeleteEntry removes a single entry by ID. Admin routes can use this to clean
// up mistaken submissions without needing direct SQL in the handler layer.
func (s *Store) DeleteEntry(ctx context.Context, id string) error {
	const query = `
		DELETE FROM entries
		WHERE id = $1
	`

	commandTag, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete entry %s: %w", id, err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func scanEntryRow(scanner interface {
	Scan(dest ...any) error
}) (Entry, error) {
	var entry Entry
	var picksRaw []byte

	err := scanner.Scan(
		&entry.ID,
		&entry.Year,
		&entry.ClerkID,
		&entry.DisplayName,
		&picksRaw,
		&entry.InOvers,
		&entry.Locked,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, ErrNotFound
		}

		return Entry{}, err
	}

	if len(picksRaw) > 0 {
		if err := json.Unmarshal(picksRaw, &entry.Picks); err != nil {
			return Entry{}, fmt.Errorf("decode entry picks json: %w", err)
		}
	}

	return entry, nil
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

	return s.getConfigByQuery(ctx, query, year)
}

// UpdateTournamentConfig updates the editable configuration row for a given year.
// Keeping this as one typed function makes the admin workflow explicit and
// keeps JSON and numeric marshaling out of the handler layer.
func (s *Store) UpdateTournamentConfig(ctx context.Context, params UpdateTournamentConfigParams) (TournamentConfig, error) {
	groupsJSON, err := json.Marshal(params.Groups)
	if err != nil {
		return TournamentConfig{}, fmt.Errorf("encode groups json: %w", err)
	}

	payoutsJSON, err := json.Marshal(params.PoolPayouts)
	if err != nil {
		return TournamentConfig{}, fmt.Errorf("encode pool payouts json: %w", err)
	}

	const query = `
		UPDATE tournament_config
		SET
			entry_deadline = $2,
			start_date = $3,
			end_date = $4,
			groups = $5::jsonb,
			mutt_multiplier = $6::numeric,
			old_mutt_multiplier = $7::numeric,
			pool_payouts = $8::jsonb,
			frl_winner = $9,
			frl_payout = $10,
			active = $11
		WHERE year = $1
		RETURNING
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
	`

	cfg, err := s.getConfigByQuery(
		ctx,
		query,
		params.Year,
		params.EntryDeadline,
		params.StartDate,
		params.EndDate,
		string(groupsJSON),
		params.MuttMultiplier,
		params.OldMuttMultiplier,
		string(payoutsJSON),
		params.FRLWinner,
		params.FRLPayout,
		params.Active,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return TournamentConfig{}, ErrNotFound
		}

		return TournamentConfig{}, fmt.Errorf("update tournament config for year %d: %w", params.Year, err)
	}

	return cfg, nil
}

func (s *Store) getConfigByQuery(ctx context.Context, query string, args ...any) (TournamentConfig, error) {
	var cfg TournamentConfig
	var groupsRaw []byte
	var payoutsRaw []byte

	err := s.pool.QueryRow(ctx, query, args...).Scan(
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

		return TournamentConfig{}, fmt.Errorf("get tournament config: %w", err)
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

func (s *Store) scanEntry(ctx context.Context, query string, args ...any) (Entry, error) {
	var entry Entry
	var picksRaw []byte

	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&entry.ID,
		&entry.Year,
		&entry.ClerkID,
		&entry.DisplayName,
		&picksRaw,
		&entry.InOvers,
		&entry.Locked,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Entry{}, ErrNotFound
		}

		return Entry{}, err
	}

	if len(picksRaw) > 0 {
		if err := json.Unmarshal(picksRaw, &entry.Picks); err != nil {
			return Entry{}, fmt.Errorf("decode entry picks json: %w", err)
		}
	}

	return entry, nil
}
