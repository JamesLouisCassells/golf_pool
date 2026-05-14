package golf

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

type Provider interface {
	FetchLeaderboard(ctx context.Context, request FetchRequest) ([]db.GolferResult, error)
}

type FetchRequest struct {
	Year         int
	TournamentID string
	RoundID      *int
}

type ProviderConfig struct {
	Provider string
	BaseURL  string
	APIKey   string
	APIHost  string
}

type rapidLeaderboardClient struct {
	baseURL    string
	apiKey     string
	apiHost    string
	httpClient *http.Client
}

func NewProvider(cfg ProviderConfig) Provider {
	if strings.TrimSpace(strings.ToLower(cfg.Provider)) != "slashgolf" {
		return nil
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://live-golf-data.p.rapidapi.com"
	}

	return &rapidLeaderboardClient{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(cfg.APIKey),
		apiHost: strings.TrimSpace(cfg.APIHost),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *rapidLeaderboardClient) FetchLeaderboard(ctx context.Context, request FetchRequest) ([]db.GolferResult, error) {
	if request.Year == 0 {
		return nil, fmt.Errorf("year is required")
	}

	tournamentID := strings.TrimSpace(request.TournamentID)
	if tournamentID == "" {
		return nil, fmt.Errorf("tournament id is required")
	}

	endpoint, err := url.Parse(c.baseURL + "/leaderboard")
	if err != nil {
		return nil, fmt.Errorf("build leaderboard endpoint: %w", err)
	}

	query := endpoint.Query()
	query.Set("tournId", tournamentID)
	query.Set("year", strconv.Itoa(request.Year))
	if request.RoundID != nil && *request.RoundID > 0 {
		query.Set("roundId", strconv.Itoa(*request.RoundID))
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build leaderboard request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-rapidapi-key", c.apiKey)
	}
	if c.apiHost != "" {
		req.Header.Set("x-rapidapi-host", c.apiHost)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch leaderboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("fetch leaderboard: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode leaderboard response: %w", err)
	}

	return extractLeaderboardResults(payload, request.Year)
}

func extractLeaderboardResults(payload any, year int) ([]db.GolferResult, error) {
	rows := leaderboardRows(payload)
	results := make([]db.GolferResult, 0, len(rows))

	for _, row := range rows {
		golferName := fullName(row)
		if golferName == "" {
			golferName = firstString(row, "playerName", "player_name", "name", "player", "golfer_name")
		}
		position := firstString(row, "pos", "position", "place", "rank")
		if golferName == "" || position == "" {
			continue
		}

		results = append(results, db.GolferResult{
			Year:       year,
			GolferName: golferName,
			Position:   position,
			Score:      firstString(row, "total", "score", "toPar", "to_par"),
			Today:      firstString(row, "today", "roundScore", "round_score"),
			Thru:       firstString(row, "thru", "status", "holesCompleted", "holes_completed"),
		})
	}

	return results, nil
}

func leaderboardRows(payload any) []map[string]any {
	switch typed := payload.(type) {
	case map[string]any:
		for _, key := range []string{"leaderboardRows", "leaderboard", "players", "results", "data"} {
			if rows, ok := typed[key].([]any); ok {
				return objectRows(rows)
			}
		}
	case []any:
		return objectRows(typed)
	}

	return nil
}

func objectRows(rows []any) []map[string]any {
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		if typed, ok := row.(map[string]any); ok {
			result = append(result, typed)
		}
	}

	return result
}

func firstString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := row[key]
		if !ok {
			continue
		}

		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		case float64:
			return strconv.FormatFloat(typed, 'f', -1, 64)
		case int:
			return strconv.Itoa(typed)
		}
	}

	return ""
}

func fullName(row map[string]any) string {
	first := firstString(row, "firstName", "first_name")
	last := firstString(row, "lastName", "last_name")
	return strings.TrimSpace(strings.Join([]string{first, last}, " "))
}
