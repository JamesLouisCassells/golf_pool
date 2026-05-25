package golf

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

type stubRefreshStore struct {
	getActiveConfigFn      func(ctx context.Context) (db.TournamentConfig, error)
	replaceGolferResultsFn func(ctx context.Context, year int, results []db.GolferResult) error
}

func (s stubRefreshStore) GetActiveConfig(ctx context.Context) (db.TournamentConfig, error) {
	return s.getActiveConfigFn(ctx)
}

func (s stubRefreshStore) ReplaceGolferResults(ctx context.Context, year int, results []db.GolferResult) error {
	return s.replaceGolferResultsFn(ctx, year, results)
}

type stubSchedulerProvider struct {
	fetchLeaderboardFn func(ctx context.Context, request FetchRequest) ([]db.GolferResult, error)
}

func (s stubSchedulerProvider) FetchLeaderboard(ctx context.Context, request FetchRequest) ([]db.GolferResult, error) {
	return s.fetchLeaderboardFn(ctx, request)
}

func TestShouldAutoRefreshOnlyDuringTournamentWindow(t *testing.T) {
	t.Parallel()

	tournamentID := "033"
	start := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	cfg := db.TournamentConfig{
		Year:                 2026,
		StartDate:            &start,
		EndDate:              &end,
		ProviderTournamentID: &tournamentID,
		Active:               true,
	}

	if shouldAutoRefresh(cfg, time.Date(2026, 5, 13, 23, 59, 0, 0, time.UTC)) {
		t.Fatalf("expected auto refresh to stay off before the tournament starts")
	}
	if !shouldAutoRefresh(cfg, time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected auto refresh to run during the tournament window")
	}
	if shouldAutoRefresh(cfg, time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected auto refresh to stop after the tournament window")
	}
}

func TestShouldAutoRefreshRequiresProviderTournamentID(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	cfg := db.TournamentConfig{
		Year:      2026,
		StartDate: &start,
		EndDate:   &end,
		Active:    true,
	}

	if shouldAutoRefresh(cfg, time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected auto refresh to stay off when provider tournament id is missing")
	}
}

func TestSchedulerTickFetchesAndStoresDuringActiveWindow(t *testing.T) {
	t.Parallel()

	tournamentID := "033"
	start := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)

	store := stubRefreshStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{
				Year:                 2026,
				StartDate:            &start,
				EndDate:              &end,
				ProviderTournamentID: &tournamentID,
				Active:               true,
			}, nil
		},
		replaceGolferResultsFn: func(ctx context.Context, year int, results []db.GolferResult) error {
			if year != 2026 {
				t.Fatalf("expected year 2026, got %d", year)
			}
			if len(results) != 1 || results[0].GolferName != "Scottie Scheffler" {
				t.Fatalf("unexpected results stored: %#v", results)
			}
			return nil
		},
	}

	provider := stubSchedulerProvider{
		fetchLeaderboardFn: func(ctx context.Context, request FetchRequest) ([]db.GolferResult, error) {
			if request.Year != 2026 {
				t.Fatalf("expected year 2026, got %d", request.Year)
			}
			if request.TournamentID != "033" {
				t.Fatalf("expected tournament id 033, got %s", request.TournamentID)
			}
			return []db.GolferResult{{Year: 2026, GolferName: "Scottie Scheffler", Position: "T1"}}, nil
		},
	}

	scheduler := NewScheduler(store, provider, log.New(io.Discard, "", 0))
	scheduler.now = func() time.Time { return now }

	if interval := scheduler.tick(context.Background()); interval != scheduler.activeInterval {
		t.Fatalf("expected next interval %s, got %s", scheduler.activeInterval, interval)
	}
}

func TestSchedulerTickBacksOffOnProviderFailure(t *testing.T) {
	t.Parallel()

	tournamentID := "033"
	start := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 18, 0, 0, 0, 0, time.UTC)

	store := stubRefreshStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{
				Year:                 2026,
				StartDate:            &start,
				EndDate:              &end,
				ProviderTournamentID: &tournamentID,
				Active:               true,
			}, nil
		},
		replaceGolferResultsFn: func(ctx context.Context, year int, results []db.GolferResult) error {
			t.Fatalf("did not expect results to be stored when fetch fails")
			return nil
		},
	}

	provider := stubSchedulerProvider{
		fetchLeaderboardFn: func(ctx context.Context, request FetchRequest) ([]db.GolferResult, error) {
			return nil, errors.New("boom")
		},
	}

	scheduler := NewScheduler(store, provider, log.New(io.Discard, "", 0))
	scheduler.now = func() time.Time { return time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC) }

	if interval := scheduler.tick(context.Background()); interval != scheduler.errorInterval {
		t.Fatalf("expected error retry interval %s, got %s", scheduler.errorInterval, interval)
	}
}
