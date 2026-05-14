package golf

import (
	"testing"
	"time"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

func TestBuildStandingsAppliesTieSplitsMultipliersAndFRLBonus(t *testing.T) {
	t.Parallel()

	frlWinner := "Scottie Scheffler"
	cfg := db.TournamentConfig{
		Year:              2026,
		PoolPayouts:       map[string]any{"1": 1000, "2": 600, "3": 400, "4": 200},
		FRLWinner:         &frlWinner,
		FRLPayout:         500000,
		MuttMultiplier:    "2",
		OldMuttMultiplier: "3",
	}

	now := time.Now().UTC()
	results := []db.GolferResult{
		{Year: 2026, GolferName: "Scottie Scheffler", Position: "T1", UpdatedAt: now},
		{Year: 2026, GolferName: "Rory McIlroy", Position: "T1", UpdatedAt: now},
		{Year: 2026, GolferName: "Justin Rose", Position: "3", UpdatedAt: now},
	}

	entries := []db.Entry{
		{
			ID:          "entry-1",
			DisplayName: "Alpha",
			Picks: map[string]any{
				"Group 1":  "Scottie Scheffler",
				"Mutt":     "Justin Rose",
				"Old Mutt": "Rory McIlroy",
			},
		},
		{
			ID:          "entry-2",
			DisplayName: "Bravo",
			Picks: map[string]any{
				"Group 1": "Rory McIlroy",
			},
		},
	}

	standings, err := BuildStandings(cfg, entries, results)
	if err != nil {
		t.Fatalf("BuildStandings returned error: %v", err)
	}

	if len(standings.Entries) != 2 {
		t.Fatalf("expected 2 standings entries, got %d", len(standings.Entries))
	}

	alpha := standings.Entries[0]
	if alpha.DisplayName != "Alpha" {
		t.Fatalf("expected Alpha to lead, got %s", alpha.DisplayName)
	}

	// T1 split = (1000 + 600) / 2 = 800
	// Scottie = 800
	// Justin Rose mutt = 400 * 2 = 800
	// Rory old mutt = 800 * 3 = 2400
	// FRL bonus = 500000
	expected := int64(504000)
	if alpha.TotalPayout != expected {
		t.Fatalf("expected Alpha total payout %d, got %d", expected, alpha.TotalPayout)
	}

	if alpha.FRLBonus != 500000 {
		t.Fatalf("expected FRL bonus 500000, got %d", alpha.FRLBonus)
	}

	bravo := standings.Entries[1]
	if bravo.DisplayName != "Bravo" {
		t.Fatalf("expected Bravo second, got %s", bravo.DisplayName)
	}

	if bravo.TotalPayout != 800 {
		t.Fatalf("expected Bravo total payout 800, got %d", bravo.TotalPayout)
	}
}

func TestBuildStandingsMatchesShortPickNamesToFullLeaderboardNames(t *testing.T) {
	t.Parallel()

	cfg := db.TournamentConfig{
		Year:        2026,
		PoolPayouts: map[string]any{"1": 1000, "2": 600, "3": 400},
	}

	results := []db.GolferResult{
		{Year: 2026, GolferName: "Scottie Scheffler", Position: "1"},
		{Year: 2026, GolferName: "Justin Rose", Position: "2"},
	}

	entries := []db.Entry{
		{
			ID:          "entry-1",
			DisplayName: "Short Names",
			Picks: map[string]any{
				"Group 1": "Scheffler",
				"WC":      "Rose",
			},
		},
	}

	standings, err := BuildStandings(cfg, entries, results)
	if err != nil {
		t.Fatalf("BuildStandings returned error: %v", err)
	}

	if len(standings.Entries) != 1 {
		t.Fatalf("expected 1 standings entry, got %d", len(standings.Entries))
	}

	if standings.Entries[0].TotalPayout != 1600 {
		t.Fatalf("expected short-name picks to total 1600, got %d", standings.Entries[0].TotalPayout)
	}
}
