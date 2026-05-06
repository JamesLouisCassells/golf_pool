package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateClaimsRejectsUnexpectedAuthorizedParty(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		AuthorizedParties: []string{"http://localhost:5173"},
	})

	err := middleware.validateClaims(tokenClaims{
		AZP:       "http://localhost:3000",
		Expiry:    numericDate(nowPlusSeconds(60)),
		NotBefore: ptrNumericDate(nowPlusSeconds(-60)),
	})
	if err == nil || err.Error() != "token authorized party mismatch" {
		t.Fatalf("expected authorized party mismatch, got %v", err)
	}
}

func TestResolveProfileUsesTokenClaimsWhenPresent(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		EmailClaim: "email",
		NameClaim:  "name",
	})

	email, displayName, err := middleware.resolveProfile(context.Background(), tokenClaims{
		Raw: map[string]any{
			"email": "james@example.com",
			"name":  "James Cassells",
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if email != "james@example.com" {
		t.Fatalf("expected email james@example.com, got %s", email)
	}
	if displayName == nil || *displayName != "James Cassells" {
		t.Fatalf("expected display name James Cassells, got %#v", displayName)
	}
}

func TestResolveProfileUsesNestedTokenClaimsWhenPresent(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		EmailClaim: "user.primary_email",
		NameClaim:  "user.display_name",
	})

	email, displayName, err := middleware.resolveProfile(context.Background(), tokenClaims{
		Raw: map[string]any{
			"user": map[string]any{
				"primary_email": "james@example.com",
				"display_name":  "James Cassells",
			},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if email != "james@example.com" {
		t.Fatalf("expected email james@example.com, got %s", email)
	}
	if displayName == nil || *displayName != "James Cassells" {
		t.Fatalf("expected display name James Cassells, got %#v", displayName)
	}
}

func TestResolveProfileFallsBackToClerkUserLookup(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		SecretKey: "sk_test_123",
		NameClaim: "name",
	})
	middleware.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/v1/users/user_123" {
				t.Fatalf("unexpected path %s", req.URL.Path)
			}
			if got := req.Header.Get("Authorization"); got != "Bearer sk_test_123" {
				t.Fatalf("unexpected authorization header %q", got)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body: io.NopCloser(strings.NewReader(`{
					"id": "user_123",
					"first_name": "James",
					"last_name": "Cassells",
					"primary_email_address_id": "idn_123",
					"email_addresses": [
						{"id": "idn_123", "email_address": "james@example.com"}
					]
				}`)),
			}, nil
		}),
	}

	original := clerkAPIBaseURL
	clerkAPIBaseURL = "https://clerk.example.test/v1"
	defer func() {
		clerkAPIBaseURL = original
	}()

	email, displayName, err := middleware.resolveProfile(context.Background(), tokenClaims{
		Subject: "user_123",
		Raw:     map[string]any{},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if email != "james@example.com" {
		t.Fatalf("expected email james@example.com, got %s", email)
	}
	if displayName == nil || *displayName != "James Cassells" {
		t.Fatalf("expected display name James Cassells, got %#v", displayName)
	}
}

func TestIsAdminRecognizesConfiguredClaim(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		AdminClaim: "role",
		AdminValue: "admin",
	})

	if !middleware.isAdmin(map[string]any{"role": "admin"}) {
		t.Fatalf("expected admin role to be recognized")
	}
}

func TestIsAdminRecognizesNestedBooleanClaim(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		AdminClaim: "app.is_admin",
		AdminValue: "true",
	})

	if !middleware.isAdmin(map[string]any{
		"app": map[string]any{
			"is_admin": true,
		},
	}) {
		t.Fatalf("expected nested boolean admin flag to be recognized")
	}
}

func TestSessionTokenFromRequestFallsBackToSessionCookie(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "__session", Value: "cookie-token"})

	if token := sessionTokenFromRequest(req); token != "cookie-token" {
		t.Fatalf("expected cookie token, got %q", token)
	}
}

func TestRequireAuthUsesMockUserWhenEnabled(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(nil, Config{
		MockEnabled: true,
		MockClerkID: "dev-user",
		MockEmail:   "dev@example.com",
		MockName:    "Dev User",
		MockAdmin:   true,
	})

	handler := middleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUser(r.Context())
		if !ok {
			t.Fatalf("expected user in context")
		}
		if user.Record.ClerkID != "dev-user" {
			t.Fatalf("expected dev-user, got %s", user.Record.ClerkID)
		}
		if !user.IsAdmin {
			t.Fatalf("expected mock user to be admin")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func nowPlusSeconds(seconds int64) time.Time {
	return time.Now().UTC().Add(time.Duration(seconds) * time.Second)
}

func ptrNumericDate(value time.Time) *numericDate {
	date := numericDate(value)
	return &date
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
