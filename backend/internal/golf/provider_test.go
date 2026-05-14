package golf

import (
	"testing"
)

func TestExtractLeaderboardResultsAcceptsCommonSlashGolfFields(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"leaderboard": []any{
			map[string]any{
				"playerName": "Scottie Scheffler",
				"pos":        "T1",
				"total":      "-12",
				"today":      "-4",
				"thru":       "F",
			},
		},
	}

	results, err := extractLeaderboardResults(payload, 2026)
	if err != nil {
		t.Fatalf("extractLeaderboardResults returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 golfer result, got %d", len(results))
	}

	if results[0].GolferName != "Scottie Scheffler" {
		t.Fatalf("expected golfer name Scottie Scheffler, got %s", results[0].GolferName)
	}

	if results[0].Position != "T1" {
		t.Fatalf("expected position T1, got %s", results[0].Position)
	}
}
