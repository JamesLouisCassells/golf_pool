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
	getConfigFn       func(ctx context.Context, year int) (db.TournamentConfig, error)
	getActiveConfigFn func(ctx context.Context) (db.TournamentConfig, error)
	getMyEntryFn      func(ctx context.Context, clerkID string) (db.Entry, error)
	listEntriesFn     func(ctx context.Context) ([]db.Entry, error)
	createEntryFn     func(ctx context.Context, params db.CreateEntryParams) (db.Entry, error)
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

func (s stubStore) ListEntriesForActiveYear(ctx context.Context) ([]db.Entry, error) {
	if s.listEntriesFn == nil {
		return nil, errors.New("unexpected ListEntriesForActiveYear call")
	}

	return s.listEntriesFn(ctx)
}

func (s stubStore) CreateEntry(ctx context.Context, params db.CreateEntryParams) (db.Entry, error) {
	if s.createEntryFn == nil {
		return db.Entry{}, errors.New("unexpected CreateEntry call")
	}

	return s.createEntryFn(ctx, params)
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
