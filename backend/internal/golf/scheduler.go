package golf

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

const (
	defaultActiveRefreshInterval = 5 * time.Minute
	defaultIdleRefreshInterval   = 6 * time.Hour
	defaultErrorRetryInterval    = 15 * time.Minute
)

type refreshStore interface {
	GetActiveConfig(ctx context.Context) (db.TournamentConfig, error)
	ReplaceGolferResults(ctx context.Context, year int, results []db.GolferResult) error
}

type Scheduler struct {
	store          refreshStore
	provider       Provider
	logger         *log.Logger
	now            func() time.Time
	activeInterval time.Duration
	idleInterval   time.Duration
	errorInterval  time.Duration
}

func NewScheduler(store refreshStore, provider Provider, logger *log.Logger) *Scheduler {
	if store == nil || provider == nil {
		return nil
	}
	if logger == nil {
		logger = log.Default()
	}

	return &Scheduler{
		store:          store,
		provider:       provider,
		logger:         logger,
		now:            time.Now,
		activeInterval: defaultActiveRefreshInterval,
		idleInterval:   defaultIdleRefreshInterval,
		errorInterval:  defaultErrorRetryInterval,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	if s == nil {
		return
	}

	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		nextInterval := s.tick(ctx)
		if nextInterval <= 0 {
			nextInterval = s.idleInterval
		}
		timer.Reset(nextInterval)
	}
}

func (s *Scheduler) tick(ctx context.Context) time.Duration {
	cfg, err := s.store.GetActiveConfig(ctx)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return s.idleInterval
		}

		s.logger.Printf("auto refresh: failed to load active config: %v", err)
		return s.errorInterval
	}

	now := s.now().UTC()
	if !shouldAutoRefresh(cfg, now) {
		return s.idleInterval
	}

	tournamentID := providerTournamentID(cfg)
	results, err := s.provider.FetchLeaderboard(ctx, FetchRequest{
		Year:         cfg.Year,
		TournamentID: tournamentID,
	})
	if err != nil {
		s.logger.Printf("auto refresh: failed to fetch leaderboard for %d/%s: %v", cfg.Year, tournamentID, err)
		return s.errorInterval
	}

	if err := s.store.ReplaceGolferResults(ctx, cfg.Year, results); err != nil {
		s.logger.Printf("auto refresh: failed to store leaderboard for %d/%s: %v", cfg.Year, tournamentID, err)
		return s.errorInterval
	}

	s.logger.Printf("auto refresh: stored %d golfer results for %d/%s", len(results), cfg.Year, tournamentID)
	return s.activeInterval
}

func shouldAutoRefresh(cfg db.TournamentConfig, now time.Time) bool {
	if providerTournamentID(cfg) == "" {
		return false
	}
	if cfg.StartDate == nil || cfg.EndDate == nil {
		return false
	}

	return withinTournamentWindow(*cfg.StartDate, *cfg.EndDate, now.UTC())
}

func providerTournamentID(cfg db.TournamentConfig) string {
	if cfg.ProviderTournamentID == nil {
		return ""
	}

	return strings.TrimSpace(*cfg.ProviderTournamentID)
}

func withinTournamentWindow(startDate, endDate time.Time, now time.Time) bool {
	startOfDay := time.Date(startDate.UTC().Year(), startDate.UTC().Month(), startDate.UTC().Day(), 0, 0, 0, 0, time.UTC)
	endExclusive := time.Date(endDate.UTC().Year(), endDate.UTC().Month(), endDate.UTC().Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)

	return !now.Before(startOfDay) && now.Before(endExclusive)
}
