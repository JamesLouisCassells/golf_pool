package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/auth"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

type stubStore struct {
	getConfigFn            func(ctx context.Context, year int) (db.TournamentConfig, error)
	getActiveConfigFn      func(ctx context.Context) (db.TournamentConfig, error)
	getMyEntryFn           func(ctx context.Context, clerkID string) (db.Entry, error)
	getEntryByIDFn         func(ctx context.Context, id string) (db.Entry, error)
	listEntriesFn          func(ctx context.Context) ([]db.Entry, error)
	listEntriesForYearFn   func(ctx context.Context, year int) ([]db.Entry, error)
	createEntryFn          func(ctx context.Context, params db.CreateEntryParams) (db.Entry, error)
	updateEntryFn          func(ctx context.Context, params db.UpdateEntryParams) (db.Entry, error)
	deleteEntryFn          func(ctx context.Context, id string) error
	updateConfigFn         func(ctx context.Context, params db.UpdateTournamentConfigParams) (db.TournamentConfig, error)
	listGolferResultsFn    func(ctx context.Context, year int) ([]db.GolferResult, error)
	replaceGolferResultsFn func(ctx context.Context, year int, results []db.GolferResult) error
	lockActiveEntriesFn    func(ctx context.Context, lockedAt time.Time) (db.LockEntriesResult, error)
}

func (s stubStore) GetConfig(ctx context.Context, year int) (db.TournamentConfig, error) {
	if s.getConfigFn == nil {
		return db.TournamentConfig{}, errors.New("unexpected GetConfig call")
	}

	return s.getConfigFn(ctx, year)
}

func (s stubStore) GetActiveConfig(ctx context.Context) (db.TournamentConfig, error) {
	if s.getActiveConfigFn == nil {
		return db.TournamentConfig{}, errors.New("unexpected GetActiveConfig call")
	}

	return s.getActiveConfigFn(ctx)
}

func (s stubStore) GetMyEntry(ctx context.Context, clerkID string) (db.Entry, error) {
	if s.getMyEntryFn == nil {
		return db.Entry{}, errors.New("unexpected GetMyEntry call")
	}

	return s.getMyEntryFn(ctx, clerkID)
}

func (s stubStore) GetEntryByID(ctx context.Context, id string) (db.Entry, error) {
	if s.getEntryByIDFn == nil {
		return db.Entry{}, errors.New("unexpected GetEntryByID call")
	}

	return s.getEntryByIDFn(ctx, id)
}

func (s stubStore) ListEntriesForActiveYear(ctx context.Context) ([]db.Entry, error) {
	if s.listEntriesFn == nil {
		return nil, errors.New("unexpected ListEntriesForActiveYear call")
	}

	return s.listEntriesFn(ctx)
}

func (s stubStore) ListEntriesForYear(ctx context.Context, year int) ([]db.Entry, error) {
	if s.listEntriesForYearFn == nil {
		return nil, errors.New("unexpected ListEntriesForYear call")
	}

	return s.listEntriesForYearFn(ctx, year)
}

func (s stubStore) CreateEntry(ctx context.Context, params db.CreateEntryParams) (db.Entry, error) {
	if s.createEntryFn == nil {
		return db.Entry{}, errors.New("unexpected CreateEntry call")
	}

	return s.createEntryFn(ctx, params)
}

func (s stubStore) UpdateEntry(ctx context.Context, params db.UpdateEntryParams) (db.Entry, error) {
	if s.updateEntryFn == nil {
		return db.Entry{}, errors.New("unexpected UpdateEntry call")
	}

	return s.updateEntryFn(ctx, params)
}

func (s stubStore) DeleteEntry(ctx context.Context, id string) error {
	if s.deleteEntryFn == nil {
		return errors.New("unexpected DeleteEntry call")
	}

	return s.deleteEntryFn(ctx, id)
}

func (s stubStore) UpdateTournamentConfig(ctx context.Context, params db.UpdateTournamentConfigParams) (db.TournamentConfig, error) {
	if s.updateConfigFn == nil {
		return db.TournamentConfig{}, errors.New("unexpected UpdateTournamentConfig call")
	}

	return s.updateConfigFn(ctx, params)
}

func (s stubStore) ListGolferResults(ctx context.Context, year int) ([]db.GolferResult, error) {
	if s.listGolferResultsFn == nil {
		return nil, errors.New("unexpected ListGolferResults call")
	}

	return s.listGolferResultsFn(ctx, year)
}

func (s stubStore) ReplaceGolferResults(ctx context.Context, year int, results []db.GolferResult) error {
	if s.replaceGolferResultsFn == nil {
		return errors.New("unexpected ReplaceGolferResults call")
	}

	return s.replaceGolferResultsFn(ctx, year, results)
}

func (s stubStore) LockActiveEntries(ctx context.Context, lockedAt time.Time) (db.LockEntriesResult, error) {
	if s.lockActiveEntriesFn == nil {
		return db.LockEntriesResult{}, errors.New("unexpected LockActiveEntries call")
	}

	return s.lockActiveEntriesFn(ctx, lockedAt)
}

func TestProtectedMeRouteRequiresBearerToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, auth.NewMiddleware(nil, auth.Config{}))

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestProtectedMyEntryRouteRequiresBearerToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, auth.NewMiddleware(nil, auth.Config{}))

	req := httptest.NewRequest(http.MethodGet, "/api/entries/mine", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestProtectedCreateEntryRouteRequiresBearerToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, auth.NewMiddleware(nil, auth.Config{}))

	req := httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestProtectedUpdateEntryRouteRequiresBearerToken(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, auth.NewMiddleware(nil, auth.Config{}))

	req := httptest.NewRequest(http.MethodPut, "/api/entries/entry-1", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestAdminConfigRouteReturnsForbiddenForNonAdmin(t *testing.T) {
	t.Parallel()

	router := NewRouter(stubStore{}, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "dev-user",
		MockEmail:   "dev@example.com",
		MockAdmin:   false,
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/config/2026", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestAdminConfigRouteReturnsConfigForAdmin(t *testing.T) {
	t.Parallel()

	store := stubStore{
		getConfigFn: func(ctx context.Context, year int) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: year, Active: true}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/config/2026", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAdminConfigUpdateSucceedsForAdmin(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	store := stubStore{
		updateConfigFn: func(ctx context.Context, params db.UpdateTournamentConfigParams) (db.TournamentConfig, error) {
			if params.Year != 2026 {
				t.Fatalf("expected year 2026, got %d", params.Year)
			}
			if params.MuttMultiplier != "2.5" {
				t.Fatalf("expected mutt multiplier 2.5, got %s", params.MuttMultiplier)
			}
			return db.TournamentConfig{
				Year:              params.Year,
				EntryDeadline:     params.EntryDeadline,
				StartDate:         params.StartDate,
				EndDate:           params.EndDate,
				Groups:            params.Groups,
				MuttMultiplier:    params.MuttMultiplier,
				OldMuttMultiplier: params.OldMuttMultiplier,
				PoolPayouts:       params.PoolPayouts,
				FRLWinner:         params.FRLWinner,
				FRLPayout:         params.FRLPayout,
				Active:            params.Active,
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	reqBody := `{"entry_deadline":"` + now.Format(time.RFC3339) + `","start_date":"` + now.Format(time.RFC3339) + `","end_date":"` + now.Format(time.RFC3339) + `","groups":{"Group 1":["Scheffler"]},"mutt_multiplier":"2.5","old_mutt_multiplier":"3.5","pool_payouts":{"1":4475},"frl_winner":"Rose","frl_payout":500000,"active":true}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/config/2026", bytes.NewBufferString(reqBody))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAdminEntriesRouteReturnsForbiddenForNonAdmin(t *testing.T) {
	t.Parallel()

	router := NewRouter(stubStore{}, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "dev-user",
		MockEmail:   "dev@example.com",
		MockAdmin:   false,
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/entries", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestAdminEntriesRouteReturnsEntriesForAdmin(t *testing.T) {
	t.Parallel()

	store := stubStore{
		listEntriesFn: func(ctx context.Context) ([]db.Entry, error) {
			return []db.Entry{
				{ID: "entry-1", Year: 2026, DisplayName: "James", Picks: map[string]any{"Group 1": "Scheffler"}},
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/entries", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAdminEntryUpdateSucceedsForAdmin(t *testing.T) {
	t.Parallel()

	store := stubStore{
		getEntryByIDFn: func(ctx context.Context, id string) (db.Entry, error) {
			return db.Entry{
				ID:          id,
				Year:        2026,
				DisplayName: "Existing Name",
				Picks:       map[string]any{"Group 1": "Spieth"},
			}, nil
		},
		updateEntryFn: func(ctx context.Context, params db.UpdateEntryParams) (db.Entry, error) {
			if params.ID != "entry-1" {
				t.Fatalf("expected entry id entry-1, got %s", params.ID)
			}
			if params.DisplayName != "James" {
				t.Fatalf("expected display name James, got %s", params.DisplayName)
			}

			return db.Entry{
				ID:          params.ID,
				Year:        2026,
				DisplayName: params.DisplayName,
				Picks:       params.Picks,
				InOvers:     params.InOvers,
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	req := httptest.NewRequest(http.MethodPut, "/api/admin/entries/entry-1", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"},"in_overs":true}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAdminEntryDeleteReturnsNoContentForAdmin(t *testing.T) {
	t.Parallel()

	store := stubStore{
		deleteEntryFn: func(ctx context.Context, id string) error {
			if id != "entry-1" {
				t.Fatalf("expected entry id entry-1, got %s", id)
			}

			return nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/entries/entry-1", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestListEntriesReturnsForbiddenBeforeTournamentStarts(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(2 * time.Hour)
	store := stubStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, StartDate: &future, Active: true}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{}))
	req := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestListEntriesReturnsEntriesAfterTournamentStarts(t *testing.T) {
	t.Parallel()

	past := time.Now().UTC().Add(-2 * time.Hour)
	store := stubStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, StartDate: &past, Active: true}, nil
		},
		listEntriesFn: func(ctx context.Context) ([]db.Entry, error) {
			return []db.Entry{
				{
					ID:          "entry-1",
					Year:        2026,
					DisplayName: "James",
					Picks:       map[string]any{"Group 1": "Scheffler"},
					CreatedAt:   time.Unix(1, 0).UTC(),
					UpdatedAt:   time.Unix(2, 0).UTC(),
				},
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{}))
	req := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestGetMyEntryReturnsEntryForAuthenticatedUser(t *testing.T) {
	t.Parallel()

	store := stubStore{
		getMyEntryFn: func(ctx context.Context, clerkID string) (db.Entry, error) {
			if clerkID != "clerk_123" {
				t.Fatalf("expected clerk ID clerk_123, got %s", clerkID)
			}

			return db.Entry{
				ID:          "entry-1",
				Year:        2026,
				DisplayName: "James",
				Picks: map[string]any{
					"Group 1": "Scheffler",
				},
				CreatedAt: time.Unix(1, 0).UTC(),
				UpdatedAt: time.Unix(2, 0).UTC(),
			}, nil
		},
	}

	handler := Handler{store: store}
	req := httptest.NewRequest(http.MethodGet, "/api/entries/mine", nil)
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{
		Record: db.User{ClerkID: "clerk_123", Email: "james@example.com"},
	}))
	recorder := httptest.NewRecorder()

	handler.getMyEntry(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestGetMyEntryReturnsNotFoundWhenNoActiveEntryExists(t *testing.T) {
	t.Parallel()

	store := stubStore{
		getMyEntryFn: func(ctx context.Context, clerkID string) (db.Entry, error) {
			return db.Entry{}, db.ErrNotFound
		},
	}

	handler := Handler{store: store}
	req := httptest.NewRequest(http.MethodGet, "/api/entries/mine", nil)
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{
		Record: db.User{ClerkID: "clerk_123", Email: "james@example.com"},
	}))
	recorder := httptest.NewRecorder()

	handler.getMyEntry(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestStandingsRouteReturnsProjectedTotals(t *testing.T) {
	t.Parallel()

	frlWinner := "Scottie Scheffler"
	store := stubStore{
		getConfigFn: func(ctx context.Context, year int) (db.TournamentConfig, error) {
			return db.TournamentConfig{
				Year:        year,
				PoolPayouts: map[string]any{"1": 1000, "2": 600},
				FRLWinner:   &frlWinner,
				FRLPayout:   500000,
			}, nil
		},
		listEntriesForYearFn: func(ctx context.Context, year int) ([]db.Entry, error) {
			return []db.Entry{
				{
					ID:          "entry-1",
					DisplayName: "James",
					Picks:       map[string]any{"Group 1": "Scottie Scheffler"},
				},
			}, nil
		},
		listGolferResultsFn: func(ctx context.Context, year int) ([]db.GolferResult, error) {
			return []db.GolferResult{
				{Year: year, GolferName: "Scottie Scheffler", Position: "1"},
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{}))
	req := httptest.NewRequest(http.MethodGet, "/api/standings/2026", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAdminRefreshSucceedsForAdmin(t *testing.T) {
	t.Parallel()

	store := stubStore{
		replaceGolferResultsFn: func(ctx context.Context, year int, results []db.GolferResult) error {
			if year != 2026 {
				t.Fatalf("expected year 2026, got %d", year)
			}
			if len(results) != 1 {
				t.Fatalf("expected 1 golfer result, got %d", len(results))
			}
			if results[0].GolferName != "Scottie Scheffler" {
				t.Fatalf("expected golfer Scottie Scheffler, got %s", results[0].GolferName)
			}
			return nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/refresh", bytes.NewBufferString(`{"year":2026,"results":[{"golfer_name":"Scottie Scheffler","position":"1","score":"-12","today":"-4","thru":"F"}]}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestAdminLockReturnsSuccessForAdmin(t *testing.T) {
	t.Parallel()

	store := stubStore{
		lockActiveEntriesFn: func(ctx context.Context, lockedAt time.Time) (db.LockEntriesResult, error) {
			return db.LockEntriesResult{
				Year:          2026,
				EntryDeadline: lockedAt,
				LockedEntries: 4,
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "admin-user",
		MockEmail:   "admin@example.com",
		MockAdmin:   true,
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/lock", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}

func TestCreateEntryReturnsNotFoundWhenNoActiveConfigExists(t *testing.T) {
	t.Parallel()

	store := stubStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{}, db.ErrNotFound
		},
	}

	handler := Handler{store: store}
	req := httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{
		Record: db.User{ClerkID: "clerk_123", Email: "james@example.com"},
	}))
	recorder := httptest.NewRecorder()

	handler.createEntry(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestCreateEntryReturnsLockedWhenDeadlineHasPassed(t *testing.T) {
	t.Parallel()

	past := time.Now().UTC().Add(-1 * time.Minute)
	store := stubStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, EntryDeadline: &past, Active: true}, nil
		},
	}

	handler := Handler{store: store}
	req := httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{
		Record: db.User{ClerkID: "clerk_123", Email: "james@example.com"},
	}))
	recorder := httptest.NewRecorder()

	handler.createEntry(recorder, req)

	if recorder.Code != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, recorder.Code)
	}
}

func TestCreateEntryReturnsConflictWhenEntryAlreadyExists(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(1 * time.Hour)
	store := stubStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, EntryDeadline: &future, Active: true}, nil
		},
		getMyEntryFn: func(ctx context.Context, clerkID string) (db.Entry, error) {
			return db.Entry{ID: "entry-1"}, nil
		},
	}

	handler := Handler{store: store}
	req := httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{
		Record: db.User{ClerkID: "clerk_123", Email: "james@example.com"},
	}))
	recorder := httptest.NewRecorder()

	handler.createEntry(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, recorder.Code)
	}
}

func TestCreateEntryCreatesEntryForAuthenticatedUser(t *testing.T) {
	t.Parallel()

	future := time.Now().UTC().Add(1 * time.Hour)
	store := stubStore{
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, EntryDeadline: &future, Active: true}, nil
		},
		getMyEntryFn: func(ctx context.Context, clerkID string) (db.Entry, error) {
			return db.Entry{}, db.ErrNotFound
		},
		createEntryFn: func(ctx context.Context, params db.CreateEntryParams) (db.Entry, error) {
			if params.Year != 2026 {
				t.Fatalf("expected year 2026, got %d", params.Year)
			}
			if params.ClerkID != "clerk_123" {
				t.Fatalf("expected clerk ID clerk_123, got %s", params.ClerkID)
			}
			if params.DisplayName != "James" {
				t.Fatalf("expected display name James, got %s", params.DisplayName)
			}

			return db.Entry{
				ID:          "entry-1",
				Year:        params.Year,
				DisplayName: params.DisplayName,
				Picks:       params.Picks,
				InOvers:     params.InOvers,
				CreatedAt:   time.Unix(1, 0).UTC(),
				UpdatedAt:   time.Unix(2, 0).UTC(),
			}, nil
		},
	}

	handler := Handler{store: store}
	req := httptest.NewRequest(http.MethodPost, "/api/entries", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"},"in_overs":true}`))
	req = req.WithContext(auth.WithUser(req.Context(), auth.User{
		Record: db.User{ClerkID: "clerk_123", Email: "james@example.com"},
	}))
	recorder := httptest.NewRecorder()

	handler.createEntry(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, recorder.Code)
	}
}

func TestUpdateEntryReturnsNotFoundWhenEntryDoesNotExist(t *testing.T) {
	t.Parallel()

	store := stubStore{
		getEntryByIDFn: func(ctx context.Context, id string) (db.Entry, error) {
			return db.Entry{}, db.ErrNotFound
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "clerk_123",
		MockEmail:   "james@example.com",
	}))
	req := httptest.NewRequest(http.MethodPut, "/api/entries/entry-1", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, recorder.Code)
	}
}

func TestUpdateEntryReturnsForbiddenWhenUserDoesNotOwnEntry(t *testing.T) {
	t.Parallel()

	otherClerkID := "other-user"
	store := stubStore{
		getEntryByIDFn: func(ctx context.Context, id string) (db.Entry, error) {
			return db.Entry{ID: "entry-1", Year: 2026, ClerkID: &otherClerkID, DisplayName: "Other"}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: "clerk_123",
		MockEmail:   "james@example.com",
	}))
	req := httptest.NewRequest(http.MethodPut, "/api/entries/entry-1", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestUpdateEntryReturnsLockedWhenDeadlineHasPassed(t *testing.T) {
	t.Parallel()

	clerkID := "clerk_123"
	past := time.Now().UTC().Add(-1 * time.Minute)
	store := stubStore{
		getEntryByIDFn: func(ctx context.Context, id string) (db.Entry, error) {
			return db.Entry{ID: id, Year: 2026, ClerkID: &clerkID, DisplayName: "James"}, nil
		},
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, EntryDeadline: &past, Active: true}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: clerkID,
		MockEmail:   "james@example.com",
	}))
	req := httptest.NewRequest(http.MethodPut, "/api/entries/entry-1", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusLocked {
		t.Fatalf("expected status %d, got %d", http.StatusLocked, recorder.Code)
	}
}

func TestUpdateEntryReturnsForbiddenWhenEntryIsNotInActiveYear(t *testing.T) {
	t.Parallel()

	clerkID := "clerk_123"
	future := time.Now().UTC().Add(1 * time.Hour)
	store := stubStore{
		getEntryByIDFn: func(ctx context.Context, id string) (db.Entry, error) {
			return db.Entry{ID: id, Year: 2025, ClerkID: &clerkID, DisplayName: "James"}, nil
		},
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, EntryDeadline: &future, Active: true}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: clerkID,
		MockEmail:   "james@example.com",
	}))
	req := httptest.NewRequest(http.MethodPut, "/api/entries/entry-1", bytes.NewBufferString(`{"display_name":"James","picks":{"Group 1":"Scheffler"}}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

func TestUpdateEntryUpdatesOwnedEntryBeforeDeadline(t *testing.T) {
	t.Parallel()

	clerkID := "clerk_123"
	future := time.Now().UTC().Add(1 * time.Hour)
	store := stubStore{
		getEntryByIDFn: func(ctx context.Context, id string) (db.Entry, error) {
			return db.Entry{
				ID:          id,
				Year:        2026,
				ClerkID:     &clerkID,
				DisplayName: "James",
				Picks:       map[string]any{"Group 1": "Old Pick"},
			}, nil
		},
		getActiveConfigFn: func(ctx context.Context) (db.TournamentConfig, error) {
			return db.TournamentConfig{Year: 2026, EntryDeadline: &future, Active: true}, nil
		},
		updateEntryFn: func(ctx context.Context, params db.UpdateEntryParams) (db.Entry, error) {
			if params.ID != "entry-1" {
				t.Fatalf("expected entry id entry-1, got %s", params.ID)
			}
			if params.DisplayName != "James Updated" {
				t.Fatalf("expected updated display name, got %s", params.DisplayName)
			}
			return db.Entry{
				ID:          params.ID,
				Year:        2026,
				ClerkID:     &clerkID,
				DisplayName: params.DisplayName,
				Picks:       params.Picks,
				InOvers:     params.InOvers,
				CreatedAt:   time.Unix(1, 0).UTC(),
				UpdatedAt:   time.Unix(3, 0).UTC(),
			}, nil
		},
	}

	router := NewRouter(store, auth.NewMiddleware(nil, auth.Config{
		MockEnabled: true,
		MockClerkID: clerkID,
		MockEmail:   "james@example.com",
	}))
	req := httptest.NewRequest(http.MethodPut, "/api/entries/entry-1", bytes.NewBufferString(`{"display_name":"James Updated","picks":{"Group 1":"Scheffler"},"in_overs":true}`))
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}
