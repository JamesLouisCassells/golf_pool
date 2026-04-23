package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store *db.Store
}

// NewRouter wires the HTTP surface for the API.
// Keeping route setup in one place makes it easier to see what the server
// exposes today and where new handlers should be added later.
func NewRouter(store *db.Store) http.Handler {
	h := Handler{store: store}
	r := chi.NewRouter()

	r.Get("/healthz", h.healthz)
	r.Get("/api/config/{year}", h.getConfig)

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
