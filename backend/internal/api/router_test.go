package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/auth"
)

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
