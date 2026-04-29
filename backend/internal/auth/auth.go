package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

type Config struct {
	JWKSURL           string
	Issuer            string
	Audience          string
	SecretKey         string
	AuthorizedParties []string
	EmailClaim        string
	NameClaim         string
	AdminClaim        string
	AdminValue        string
}

type Middleware struct {
	store      *db.Store
	config     Config
	httpClient *http.Client

	mu         sync.RWMutex
	keyCache   map[string]*rsa.PublicKey
	cacheUntil time.Time
}

type User struct {
	Record  db.User `json:"user"`
	IsAdmin bool    `json:"is_admin"`
}

type contextKey string

const userContextKey contextKey = "auth.user"

type tokenHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

type tokenClaims struct {
	Subject string `json:"sub"`

	Issuer    string        `json:"iss"`
	Audience  audienceClaim `json:"aud"`
	Expiry    numericDate   `json:"exp"`
	NotBefore *numericDate  `json:"nbf,omitempty"`
	AZP       string        `json:"azp,omitempty"`
	Status    string        `json:"sts,omitempty"`

	Raw map[string]any `json:"-"`
}

type clerkUserResponse struct {
	ID                    string              `json:"id"`
	FirstName             *string             `json:"first_name"`
	LastName              *string             `json:"last_name"`
	PrimaryEmailAddressID *string             `json:"primary_email_address_id"`
	EmailAddresses        []clerkEmailAddress `json:"email_addresses"`
}

type clerkEmailAddress struct {
	ID           string `json:"id"`
	EmailAddress string `json:"email_address"`
}

var clerkAPIBaseURL = "https://api.clerk.com/v1"

type numericDate time.Time

func (n *numericDate) UnmarshalJSON(data []byte) error {
	var seconds float64
	if err := json.Unmarshal(data, &seconds); err != nil {
		return err
	}

	*n = numericDate(time.Unix(int64(seconds), 0).UTC())
	return nil
}

func (n numericDate) Time() time.Time {
	return time.Time(n)
}

type audienceClaim []string

func (a *audienceClaim) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = audienceClaim{single}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}

	*a = audienceClaim(many)
	return nil
}

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func NewMiddleware(store *db.Store, cfg Config) *Middleware {
	return &Middleware{
		store:  store,
		config: cfg,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		keyCache: make(map[string]*rsa.PublicKey),
	}
}

// RequireAuth verifies the incoming bearer token, syncs the local user record,
// and attaches the authenticated user to the request context.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := sessionTokenFromRequest(r)
		if token == "" {
			http.Error(w, "missing session token", http.StatusUnauthorized)
			return
		}

		if m.config.JWKSURL == "" {
			http.Error(w, "auth is not configured", http.StatusServiceUnavailable)
			return
		}

		authUser, err := m.authenticate(r.Context(), token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, authUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey).(User)
	return user, ok
}

// WithUser attaches an already-authenticated user to a context. This keeps
// tests and future middleware composition from needing to duplicate the
// package's private context key details.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// RequireAdmin builds on RequireAuth and then enforces the derived admin flag.
func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUser(r.Context())
		if !ok {
			http.Error(w, "authenticated user missing from context", http.StatusInternalServerError)
			return
		}

		if !user.IsAdmin {
			http.Error(w, "admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}))
}

func (m *Middleware) authenticate(ctx context.Context, token string) (User, error) {
	header, claims, signingInput, signature, err := parseToken(token)
	if err != nil {
		return User{}, err
	}

	if header.Alg != "RS256" {
		return User{}, errors.New("unsupported jwt algorithm")
	}

	if header.Kid == "" {
		return User{}, errors.New("token is missing key id")
	}

	key, err := m.lookupKey(ctx, header.Kid)
	if err != nil {
		return User{}, err
	}

	if err := verifyRS256(signingInput, signature, key); err != nil {
		return User{}, errors.New("invalid token signature")
	}

	if err := m.validateClaims(claims); err != nil {
		return User{}, err
	}

	if claims.Subject == "" {
		return User{}, errors.New("token is missing subject")
	}

	email, displayName, err := m.resolveProfile(ctx, claims)
	if err != nil {
		return User{}, err
	}

	record, err := m.store.UpsertUser(ctx, claims.Subject, email, displayName)
	if err != nil {
		return User{}, fmt.Errorf("sync local user: %w", err)
	}

	return User{
		Record:  record,
		IsAdmin: m.isAdmin(claims.Raw),
	}, nil
}

func (m *Middleware) validateClaims(claims tokenClaims) error {
	now := time.Now().UTC()

	if claims.Expiry.Time().IsZero() || !claims.Expiry.Time().After(now) {
		return errors.New("token is expired")
	}

	if claims.NotBefore != nil && claims.NotBefore.Time().After(now) {
		return errors.New("token is not valid yet")
	}

	if m.config.Issuer != "" && claims.Issuer != m.config.Issuer {
		return errors.New("token issuer mismatch")
	}

	if m.config.Audience != "" && !claims.Audience.Contains(m.config.Audience) {
		return errors.New("token audience mismatch")
	}

	if len(m.config.AuthorizedParties) > 0 && claims.AZP != "" && !containsString(m.config.AuthorizedParties, claims.AZP) {
		return errors.New("token authorized party mismatch")
	}

	if claims.Status == "pending" {
		return errors.New("token session status is pending")
	}

	return nil
}

func (a audienceClaim) Contains(target string) bool {
	for _, candidate := range a {
		if candidate == target {
			return true
		}
	}

	return false
}

func (m *Middleware) isAdmin(claims map[string]any) bool {
	if m.config.AdminClaim == "" || m.config.AdminValue == "" {
		return false
	}

	value, ok := claims[m.config.AdminClaim]
	if !ok {
		return false
	}

	switch typed := value.(type) {
	case string:
		return typed == m.config.AdminValue
	case []any:
		for _, item := range typed {
			if text, ok := item.(string); ok && text == m.config.AdminValue {
				return true
			}
		}
	}

	return false
}

func (m *Middleware) resolveProfile(ctx context.Context, claims tokenClaims) (string, *string, error) {
	email, _ := claimString(claims.Raw, m.config.EmailClaim)
	name, _ := claimString(claims.Raw, m.config.NameClaim)
	displayName := optionalString(name)

	if email != "" {
		return email, displayName, nil
	}

	if m.config.SecretKey == "" {
		return "", nil, errors.New("token is missing email claim and CLERK_SECRET_KEY is not set for Clerk user lookup")
	}

	profile, err := m.fetchUserProfile(ctx, claims.Subject)
	if err != nil {
		return "", nil, err
	}

	email = profile.primaryEmail()
	if email == "" {
		return "", nil, errors.New("clerk user profile did not include a primary email address")
	}

	if displayName == nil {
		displayName = optionalString(profile.displayName())
	}

	return email, displayName, nil
}

func (m *Middleware) lookupKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	m.mu.RLock()
	key, ok := m.keyCache[kid]
	cacheValid := time.Now().Before(m.cacheUntil)
	m.mu.RUnlock()

	if ok && cacheValid {
		return key, nil
	}

	keys, err := m.fetchKeys(ctx)
	if err != nil {
		return nil, err
	}

	key, ok = keys[kid]
	if !ok {
		return nil, errors.New("signing key not found")
	}

	return key, nil
}

func (m *Middleware) fetchKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.config.JWKSURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build jwks request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch jwks: unexpected status %d", resp.StatusCode)
	}

	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode jwks: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(doc.Keys))
	for _, raw := range doc.Keys {
		if raw.Kty != "RSA" || raw.Kid == "" {
			continue
		}

		key, err := parseRSAPublicKey(raw.N, raw.E)
		if err != nil {
			return nil, fmt.Errorf("parse jwks key %s: %w", raw.Kid, err)
		}

		keys[raw.Kid] = key
	}

	m.mu.Lock()
	m.keyCache = keys
	m.cacheUntil = time.Now().Add(15 * time.Minute)
	m.mu.Unlock()

	return keys, nil
}

func (m *Middleware) fetchUserProfile(ctx context.Context, userID string) (clerkUserResponse, error) {
	if userID == "" {
		return clerkUserResponse{}, errors.New("cannot fetch clerk user profile without user id")
	}

	baseURL := strings.TrimRight(clerkAPIBaseURL, "/") + "/users/" + url.PathEscape(userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
	if err != nil {
		return clerkUserResponse{}, fmt.Errorf("build clerk user request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.config.SecretKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return clerkUserResponse{}, fmt.Errorf("fetch clerk user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return clerkUserResponse{}, fmt.Errorf("fetch clerk user: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var profile clerkUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return clerkUserResponse{}, fmt.Errorf("decode clerk user: %w", err)
	}

	return profile, nil
}

func parseToken(token string) (tokenHeader, tokenClaims, string, []byte, error) {
	var header tokenHeader
	var claims tokenClaims

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return header, claims, "", nil, errors.New("token must have three parts")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return header, claims, "", nil, errors.New("failed to decode token header")
	}

	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return header, claims, "", nil, errors.New("failed to parse token header")
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return header, claims, "", nil, errors.New("failed to decode token claims")
	}

	rawClaims := make(map[string]any)
	if err := json.Unmarshal(claimsBytes, &rawClaims); err != nil {
		return header, claims, "", nil, errors.New("failed to parse token claims")
	}

	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return header, claims, "", nil, errors.New("failed to map token claims")
	}
	claims.Raw = rawClaims

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return header, claims, "", nil, errors.New("failed to decode token signature")
	}

	return header, claims, parts[0] + "." + parts[1], signature, nil
}

func verifyRS256(signingInput string, signature []byte, key *rsa.PublicKey) error {
	hash := sha256.Sum256([]byte(signingInput))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], signature)
}

func parseRSAPublicKey(nValue, eValue string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nValue)
	if err != nil {
		return nil, errors.New("invalid rsa modulus")
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eValue)
	if err != nil {
		return nil, errors.New("invalid rsa exponent")
	}

	modulus := new(big.Int).SetBytes(nBytes)
	exponent := new(big.Int).SetBytes(eBytes).Int64()
	if exponent <= 0 {
		return nil, errors.New("invalid rsa exponent value")
	}

	return &rsa.PublicKey{
		N: modulus,
		E: int(exponent),
	}, nil
}

func sessionTokenFromRequest(r *http.Request) string {
	if token := bearerToken(r.Header.Get("Authorization")); token != "" {
		return token
	}

	if cookie, err := r.Cookie("__session"); err == nil {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

func bearerToken(header string) string {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func claimString(claims map[string]any, key string) (string, bool) {
	if key == "" {
		return "", false
	}

	value, ok := claims[key]
	if !ok {
		return "", false
	}

	text, ok := value.(string)
	if !ok {
		return "", false
	}

	return strings.TrimSpace(text), true
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

func (u clerkUserResponse) primaryEmail() string {
	if u.PrimaryEmailAddressID != nil {
		for _, email := range u.EmailAddresses {
			if email.ID == *u.PrimaryEmailAddressID {
				return strings.TrimSpace(email.EmailAddress)
			}
		}
	}

	for _, email := range u.EmailAddresses {
		if strings.TrimSpace(email.EmailAddress) != "" {
			return strings.TrimSpace(email.EmailAddress)
		}
	}

	return ""
}

func (u clerkUserResponse) displayName() string {
	parts := []string{}
	if u.FirstName != nil && strings.TrimSpace(*u.FirstName) != "" {
		parts = append(parts, strings.TrimSpace(*u.FirstName))
	}
	if u.LastName != nil && strings.TrimSpace(*u.LastName) != "" {
		parts = append(parts, strings.TrimSpace(*u.LastName))
	}

	return strings.Join(parts, " ")
}
