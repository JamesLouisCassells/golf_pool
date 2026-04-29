package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/auth"
	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"

	"github.com/go-chi/chi/v5"
)

type Store interface {
	GetConfig(ctx context.Context, year int) (db.TournamentConfig, error)
	GetMyEntry(ctx context.Context, clerkID string) (db.Entry, error)
}

type Handler struct {
	store Store
}

// NewRouter wires the HTTP surface for the API.
// Keeping route setup in one place makes it easier to see what the server
// exposes today and where new handlers should be added later.
func NewRouter(store Store, authMiddleware *auth.Middleware) http.Handler {
	h := Handler{store: store}
	r := chi.NewRouter()

	r.Get("/healthz", h.healthz)
	r.Get("/api/config/{year}", h.getConfig)
	r.With(authMiddleware.RequireAuth).Get("/api/me", h.me)
	r.With(authMiddleware.RequireAuth).Get("/api/entries/mine", h.getMyEntry)

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
