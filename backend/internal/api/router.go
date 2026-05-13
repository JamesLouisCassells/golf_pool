package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/auth"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/golf"

	"github.com/go-chi/chi/v5"
)

type Store interface {
	GetConfig(ctx context.Context, year int) (db.TournamentConfig, error)
	GetActiveConfig(ctx context.Context) (db.TournamentConfig, error)
	GetMyEntry(ctx context.Context, clerkID string) (db.Entry, error)
	GetEntryByID(ctx context.Context, id string) (db.Entry, error)
	ListEntriesForActiveYear(ctx context.Context) ([]db.Entry, error)
	ListEntriesForYear(ctx context.Context, year int) ([]db.Entry, error)
	CreateEntry(ctx context.Context, params db.CreateEntryParams) (db.Entry, error)
	UpdateEntry(ctx context.Context, params db.UpdateEntryParams) (db.Entry, error)
	DeleteEntry(ctx context.Context, id string) error
	UpdateTournamentConfig(ctx context.Context, params db.UpdateTournamentConfigParams) (db.TournamentConfig, error)
	ListGolferResults(ctx context.Context, year int) ([]db.GolferResult, error)
	ReplaceGolferResults(ctx context.Context, year int, results []db.GolferResult) error
	LockActiveEntries(ctx context.Context, lockedAt time.Time) (db.LockEntriesResult, error)
}

type Handler struct {
	store Store
}

type createEntryRequest struct {
	DisplayName string         `json:"display_name"`
	Picks       map[string]any `json:"picks"`
	InOvers     bool           `json:"in_overs"`
}

type updateEntryRequest = createEntryRequest

type updateTournamentConfigRequest struct {
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

type refreshResultsRequest struct {
	Year    int                   `json:"year"`
	Results []refreshResultRecord `json:"results"`
}

type refreshResultRecord struct {
	GolferName string `json:"golfer_name"`
	Position   string `json:"position"`
	Score      string `json:"score"`
	Today      string `json:"today"`
	Thru       string `json:"thru"`
}

// NewRouter wires the HTTP surface for the API.
// Keeping route setup in one place makes it easier to see what the server
// exposes today and where new handlers should be added later.
func NewRouter(store Store, authMiddleware *auth.Middleware) http.Handler {
	h := Handler{store: store}
	r := chi.NewRouter()

	r.Get("/healthz", h.healthz)
	r.Get("/api/config/{year}", h.getConfig)
	r.Get("/api/entries", h.listEntries)
	r.Get("/api/standings/{year}", h.getStandings)
	r.With(authMiddleware.RequireAuth).Get("/api/me", h.me)
	r.With(authMiddleware.RequireAuth).Get("/api/entries/mine", h.getMyEntry)
	r.With(authMiddleware.RequireAuth).Post("/api/entries", h.createEntry)
	r.With(authMiddleware.RequireAuth).Put("/api/entries/{id}", h.updateEntry)
	r.With(authMiddleware.RequireAdmin).Get("/api/admin/config/{year}", h.getAdminConfig)
	r.With(authMiddleware.RequireAdmin).Put("/api/admin/config/{year}", h.updateAdminConfig)
	r.With(authMiddleware.RequireAdmin).Get("/api/admin/entries", h.listAdminEntries)
	r.With(authMiddleware.RequireAdmin).Put("/api/admin/entries/{id}", h.updateAdminEntry)
	r.With(authMiddleware.RequireAdmin).Delete("/api/admin/entries/{id}", h.deleteAdminEntry)
	r.With(authMiddleware.RequireAdmin).Post("/api/admin/refresh", h.refreshAdminResults)
	r.With(authMiddleware.RequireAdmin).Post("/api/admin/lock", h.lockAdminEntries)

	return r
}

func (h Handler) healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		http.Error(w, "year must be a valid integer", http.StatusBadRequest)
		return
	}

	cfg, err := h.store.GetConfig(r.Context(), year)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load config", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) getAdminConfig(w http.ResponseWriter, r *http.Request) {
	h.getConfig(w, r)
}

func (h Handler) me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		http.Error(w, "authenticated user missing from context", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) getMyEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		http.Error(w, "authenticated user missing from context", http.StatusInternalServerError)
		return
	}

	entry, err := h.store.GetMyEntry(r.Context(), user.Record.ClerkID)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "entry not found for active year", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load entry", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(entry); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) listEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	activeConfig, err := h.store.GetActiveConfig(r.Context())
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "active tournament config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load active tournament config", http.StatusInternalServerError)
		return
	}

	if tournamentNotStarted(activeConfig.StartDate, time.Now().UTC()) {
		http.Error(w, "entries are hidden until the tournament starts", http.StatusForbidden)
		return
	}

	entries, err := h.store.ListEntriesForActiveYear(r.Context())
	if err != nil {
		http.Error(w, "failed to load entries", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(entries); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) getStandings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		http.Error(w, "year must be a valid integer", http.StatusBadRequest)
		return
	}

	standings, err := h.standingsForYear(r.Context(), year)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load standings", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(standings); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) createEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		http.Error(w, "authenticated user missing from context", http.StatusInternalServerError)
		return
	}

	var request createEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body must be valid json", http.StatusBadRequest)
		return
	}

	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if request.DisplayName == "" {
		if user.Record.DisplayName != nil {
			request.DisplayName = strings.TrimSpace(*user.Record.DisplayName)
		}
		if request.DisplayName == "" {
			request.DisplayName = user.Record.Email
		}
	}

	if len(request.Picks) == 0 {
		http.Error(w, "picks are required", http.StatusBadRequest)
		return
	}

	activeConfig, err := h.store.GetActiveConfig(r.Context())
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "active tournament config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load active tournament config", http.StatusInternalServerError)
		return
	}

	if deadlinePassed(activeConfig.EntryDeadline, time.Now().UTC()) {
		http.Error(w, "entry deadline has passed", http.StatusLocked)
		return
	}

	_, err = h.store.GetMyEntry(r.Context(), user.Record.ClerkID)
	if err == nil {
		http.Error(w, "entry already exists for active year", http.StatusConflict)
		return
	}
	if err != nil && err != db.ErrNotFound {
		http.Error(w, "failed to check existing entry", http.StatusInternalServerError)
		return
	}

	entry, err := h.store.CreateEntry(r.Context(), db.CreateEntryParams{
		Year:        activeConfig.Year,
		ClerkID:     user.Record.ClerkID,
		DisplayName: request.DisplayName,
		Picks:       request.Picks,
		InOvers:     request.InOvers,
	})
	if err != nil {
		if err == db.ErrConflict {
			http.Error(w, "entry already exists for active year", http.StatusConflict)
			return
		}

		http.Error(w, "failed to create entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(entry); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) updateEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, ok := auth.CurrentUser(r.Context())
	if !ok {
		http.Error(w, "authenticated user missing from context", http.StatusInternalServerError)
		return
	}

	entryID := chi.URLParam(r, "id")
	if strings.TrimSpace(entryID) == "" {
		http.Error(w, "entry id is required", http.StatusBadRequest)
		return
	}

	existingEntry, err := h.store.GetEntryByID(r.Context(), entryID)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load entry", http.StatusInternalServerError)
		return
	}

	if existingEntry.ClerkID == nil || *existingEntry.ClerkID != user.Record.ClerkID {
		http.Error(w, "you do not own this entry", http.StatusForbidden)
		return
	}

	activeConfig, err := h.store.GetActiveConfig(r.Context())
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "active tournament config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load active tournament config", http.StatusInternalServerError)
		return
	}

	if existingEntry.Year != activeConfig.Year {
		http.Error(w, "entry is not for the active tournament year", http.StatusForbidden)
		return
	}

	if deadlinePassed(activeConfig.EntryDeadline, time.Now().UTC()) {
		http.Error(w, "entry deadline has passed", http.StatusLocked)
		return
	}

	var request updateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body must be valid json", http.StatusBadRequest)
		return
	}

	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if request.DisplayName == "" {
		request.DisplayName = existingEntry.DisplayName
	}
	if len(request.Picks) == 0 {
		http.Error(w, "picks are required", http.StatusBadRequest)
		return
	}

	updatedEntry, err := h.store.UpdateEntry(r.Context(), db.UpdateEntryParams{
		ID:          existingEntry.ID,
		DisplayName: request.DisplayName,
		Picks:       request.Picks,
		InOvers:     request.InOvers,
	})
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to update entry", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(updatedEntry); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) updateAdminConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	year, err := strconv.Atoi(chi.URLParam(r, "year"))
	if err != nil {
		http.Error(w, "year must be a valid integer", http.StatusBadRequest)
		return
	}

	var request updateTournamentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body must be valid json", http.StatusBadRequest)
		return
	}

	if len(request.Groups) == 0 {
		http.Error(w, "groups are required", http.StatusBadRequest)
		return
	}

	if len(request.PoolPayouts) == 0 {
		http.Error(w, "pool payouts are required", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(request.MuttMultiplier) == "" || strings.TrimSpace(request.OldMuttMultiplier) == "" {
		http.Error(w, "mutt multipliers are required", http.StatusBadRequest)
		return
	}

	cfg, err := h.store.UpdateTournamentConfig(r.Context(), db.UpdateTournamentConfigParams{
		Year:              year,
		EntryDeadline:     request.EntryDeadline,
		StartDate:         request.StartDate,
		EndDate:           request.EndDate,
		Groups:            request.Groups,
		MuttMultiplier:    strings.TrimSpace(request.MuttMultiplier),
		OldMuttMultiplier: strings.TrimSpace(request.OldMuttMultiplier),
		PoolPayouts:       request.PoolPayouts,
		FRLWinner:         request.FRLWinner,
		FRLPayout:         request.FRLPayout,
		Active:            request.Active,
	})
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to update config", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) listAdminEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	entries, err := h.store.ListEntriesForActiveYear(r.Context())
	if err != nil {
		http.Error(w, "failed to load admin entries", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(entries); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) updateAdminEntry(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	entryID := chi.URLParam(r, "id")
	if strings.TrimSpace(entryID) == "" {
		http.Error(w, "entry id is required", http.StatusBadRequest)
		return
	}

	existingEntry, err := h.store.GetEntryByID(r.Context(), entryID)
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to load entry", http.StatusInternalServerError)
		return
	}

	var request updateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body must be valid json", http.StatusBadRequest)
		return
	}

	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if request.DisplayName == "" {
		request.DisplayName = existingEntry.DisplayName
	}
	if len(request.Picks) == 0 {
		http.Error(w, "picks are required", http.StatusBadRequest)
		return
	}

	updatedEntry, err := h.store.UpdateEntry(r.Context(), db.UpdateEntryParams{
		ID:          existingEntry.ID,
		DisplayName: request.DisplayName,
		Picks:       request.Picks,
		InOvers:     request.InOvers,
	})
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to update entry", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(updatedEntry); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) deleteAdminEntry(w http.ResponseWriter, r *http.Request) {
	entryID := chi.URLParam(r, "id")
	if strings.TrimSpace(entryID) == "" {
		http.Error(w, "entry id is required", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteEntry(r.Context(), entryID); err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "entry not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to delete entry", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) refreshAdminResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var request refreshResultsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request body must be valid json", http.StatusBadRequest)
		return
	}

	year := request.Year
	if year == 0 {
		activeConfig, err := h.store.GetActiveConfig(r.Context())
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "active tournament config not found", http.StatusNotFound)
				return
			}

			http.Error(w, "failed to load active tournament config", http.StatusInternalServerError)
			return
		}

		year = activeConfig.Year
	}

	results := make([]db.GolferResult, 0, len(request.Results))
	for _, result := range request.Results {
		golferName := strings.TrimSpace(result.GolferName)
		position := strings.TrimSpace(result.Position)
		if golferName == "" || position == "" {
			http.Error(w, "each golfer result requires golfer_name and position", http.StatusBadRequest)
			return
		}

		results = append(results, db.GolferResult{
			Year:       year,
			GolferName: golferName,
			Position:   position,
			Score:      strings.TrimSpace(result.Score),
			Today:      strings.TrimSpace(result.Today),
			Thru:       strings.TrimSpace(result.Thru),
		})
	}

	if err := h.store.ReplaceGolferResults(r.Context(), year, results); err != nil {
		http.Error(w, "failed to refresh golfer results", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"year":         year,
		"result_count": len(results),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) lockAdminEntries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	result, err := h.store.LockActiveEntries(r.Context(), time.Now().UTC())
	if err != nil {
		if err == db.ErrNotFound {
			http.Error(w, "active tournament config not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to lock active entries", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h Handler) standingsForYear(ctx context.Context, year int) (golf.Standings, error) {
	cfg, err := h.store.GetConfig(ctx, year)
	if err != nil {
		return golf.Standings{}, err
	}

	entries, err := h.store.ListEntriesForYear(ctx, year)
	if err != nil {
		return golf.Standings{}, err
	}

	results, err := h.store.ListGolferResults(ctx, year)
	if err != nil {
		return golf.Standings{}, err
	}

	return golf.BuildStandings(cfg, entries, results)
}

func deadlinePassed(deadline *time.Time, now time.Time) bool {
	return deadline != nil && !deadline.After(now)
}

func tournamentNotStarted(startDate *time.Time, now time.Time) bool {
	if startDate == nil {
		return false
	}

	return startDate.After(now)
}
