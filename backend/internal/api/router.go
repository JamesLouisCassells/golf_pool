package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter wires the HTTP surface for the API.
// Keeping route setup in one place makes it easier to see what the server
// exposes today and where new handlers should be added later.
func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/healthz", healthz)

	return r
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Print("Hit health endpoint")

	response := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
