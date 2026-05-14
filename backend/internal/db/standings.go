package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type GolferResult struct {
	Year       int       `json:"year"`
	GolferName string    `json:"golfer_name"`
	Position   string    `json:"position"`
	Score      string    `json:"score"`
	Today      string    `json:"today"`
	Thru       string    `json:"thru"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type LockEntriesResult struct {
	Year          int        `json:"year"`
	EntryDeadline time.Time  `json:"entry_deadline"`
	LockedEntries int64      `json:"locked_entries"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

// ListEntriesForYear returns every entry for a specific tournament year.
// Standings calculation needs the full field regardless of whether that year is
// currently marked active.
func (s *Store) ListEntriesForYear(ctx context.Context, year int) ([]Entry, error) {
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
		WHERE year = $1
		ORDER BY display_name ASC, created_at ASC
	`

	rows, err := s.pool.Query(ctx, query, year)
	if err != nil {
		return nil, fmt.Errorf("list entries for year %d: %w", year, err)
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
		return nil, fmt.Errorf("iterate entries for year %d: %w", year, err)
	}

	return entries, nil
}

// ListGolferResults returns the currently stored live results snapshot for a
// tournament year.
func (s *Store) ListGolferResults(ctx context.Context, year int) ([]GolferResult, error) {
	const query = `
		SELECT
			year,
			golfer_name,
			position,
			score,
			today,
			thru,
			updated_at
		FROM golfer_results
		WHERE year = $1
		ORDER BY golfer_name ASC
	`

	rows, err := s.pool.Query(ctx, query, year)
	if err != nil {
		return nil, fmt.Errorf("list golfer results for year %d: %w", year, err)
	}
	defer rows.Close()

	results := []GolferResult{}
	for rows.Next() {
		var result GolferResult
		if err := rows.Scan(
			&result.Year,
			&result.GolferName,
			&result.Position,
			&result.Score,
			&result.Today,
			&result.Thru,
			&result.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan golfer result for year %d: %w", year, err)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate golfer results for year %d: %w", year, err)
	}

	return results, nil
}

// ReplaceGolferResults swaps the stored live results snapshot for a year. This
// makes manual admin refreshes idempotent and keeps standings reads simple.
func (s *Store) ReplaceGolferResults(ctx context.Context, year int, results []GolferResult) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin golfer results transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM golfer_results WHERE year = $1`, year); err != nil {
		return fmt.Errorf("delete golfer results for year %d: %w", year, err)
	}

	if len(results) > 0 {
		const insertQuery = `
			INSERT INTO golfer_results (
				year,
				golfer_name,
				position,
				score,
				today,
				thru,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, now())
		`

		for _, result := range results {
			if _, err := tx.Exec(
				ctx,
				insertQuery,
				year,
				result.GolferName,
				result.Position,
				result.Score,
				result.Today,
				result.Thru,
			); err != nil {
				return fmt.Errorf("insert golfer result %s for year %d: %w", result.GolferName, year, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit golfer results for year %d: %w", year, err)
	}

	return nil
}

// LockActiveEntries closes the active tournament immediately by pulling the
// deadline to now and marking all entries for that year locked.
func (s *Store) LockActiveEntries(ctx context.Context, lockedAt time.Time) (LockEntriesResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return LockEntriesResult{}, fmt.Errorf("begin lock active entries transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var year int
	err = tx.QueryRow(ctx, `
		SELECT year
		FROM tournament_config
		WHERE active = true
		ORDER BY year DESC
		LIMIT 1
	`).Scan(&year)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LockEntriesResult{}, ErrNotFound
		}

		return LockEntriesResult{}, fmt.Errorf("load active tournament year: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tournament_config
		SET entry_deadline = $2
		WHERE year = $1
	`, year, lockedAt); err != nil {
		return LockEntriesResult{}, fmt.Errorf("update active tournament deadline: %w", err)
	}

	commandTag, err := tx.Exec(ctx, `
		UPDATE entries
		SET locked = true, updated_at = now()
		WHERE year = $1
	`, year)
	if err != nil {
		return LockEntriesResult{}, fmt.Errorf("lock active entries for year %d: %w", year, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return LockEntriesResult{}, fmt.Errorf("commit lock active entries transaction: %w", err)
	}

	return LockEntriesResult{
		Year:          year,
		EntryDeadline: lockedAt,
		LockedEntries: commandTag.RowsAffected(),
	}, nil
}
