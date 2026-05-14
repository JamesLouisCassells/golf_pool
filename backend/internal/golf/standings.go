package golf

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/JamesLouisCassells/golf_pool/backend/internal/db"
)

type StandingPick struct {
	GroupName   string `json:"group_name"`
	GolferName  string `json:"golfer_name"`
	Position    string `json:"position"`
	Score       string `json:"score"`
	Today       string `json:"today"`
	Thru        string `json:"thru"`
	BasePayout  int64  `json:"base_payout"`
	Multiplier  int64  `json:"multiplier"`
	TotalPayout int64  `json:"total_payout"`
}

type StandingEntry struct {
	Rank        int            `json:"rank"`
	EntryID     string         `json:"entry_id"`
	DisplayName string         `json:"display_name"`
	InOvers     bool           `json:"in_overs"`
	TotalPayout int64          `json:"total_payout"`
	FRLBonus    int64          `json:"frl_bonus"`
	Picks       []StandingPick `json:"picks"`
}

type Standings struct {
	Year        int             `json:"year"`
	UpdatedAt   *time.Time      `json:"updated_at,omitempty"`
	Entries     []StandingEntry `json:"entries"`
	ResultCount int             `json:"result_count"`
}

func BuildStandings(cfg db.TournamentConfig, entries []db.Entry, results []db.GolferResult) (Standings, error) {
	payouts, err := parsePoolPayouts(cfg.PoolPayouts)
	if err != nil {
		return Standings{}, err
	}

	projectedByGolfer := projectedPayouts(results, payouts)
	resultByGolfer := make(map[string]db.GolferResult, len(results))
	nameAliases := buildNameAliases(results)

	var updatedAt *time.Time
	for _, result := range results {
		resultByGolfer[normalizeName(result.GolferName)] = result
		if updatedAt == nil || result.UpdatedAt.After(*updatedAt) {
			timestamp := result.UpdatedAt
			updatedAt = &timestamp
		}
	}

	standings := Standings{
		Year:        cfg.Year,
		UpdatedAt:   updatedAt,
		Entries:     make([]StandingEntry, 0, len(entries)),
		ResultCount: len(results),
	}

	frlWinnerKey := normalizeName(stringValue(cfg.FRLWinner))
	frlBonus := int64(cfg.FRLPayout)

	for _, entry := range entries {
		picks, pickedGolfers := collectPicks(entry.Picks)
		standingPicks := make([]StandingPick, 0, len(picks))
		total := int64(0)
		hasFRLWinner := false

		for _, pick := range picks {
			key := resolveGolferKey(pick.GolferName, nameAliases)
			result, found := resultByGolfer[key]
			basePayout := projectedByGolfer[key]
			multiplier := payoutMultiplier(pick.GroupName)
			totalPayout := basePayout * multiplier

			if frlWinnerKey != "" && key == resolveGolferKey(stringValue(cfg.FRLWinner), nameAliases) {
				hasFRLWinner = true
			}

			standingPick := StandingPick{
				GroupName:   pick.GroupName,
				GolferName:  pick.GolferName,
				BasePayout:  basePayout,
				Multiplier:  multiplier,
				TotalPayout: totalPayout,
			}

			if found {
				standingPick.Position = result.Position
				standingPick.Score = result.Score
				standingPick.Today = result.Today
				standingPick.Thru = result.Thru
			}

			standingPicks = append(standingPicks, standingPick)
			total += totalPayout
		}

		entryFRLBonus := int64(0)
		if frlWinnerKey != "" && hasFRLWinner {
			entryFRLBonus = frlBonus
			total += entryFRLBonus
		}

		if frlWinnerKey != "" && !hasFRLWinner {
			for _, golferName := range pickedGolfers {
				if resolveGolferKey(golferName, nameAliases) == resolveGolferKey(stringValue(cfg.FRLWinner), nameAliases) {
					entryFRLBonus = frlBonus
					total += entryFRLBonus
					break
				}
			}
		}

		sort.Slice(standingPicks, func(i, j int) bool {
			return standingPicks[i].GroupName < standingPicks[j].GroupName
		})

		standings.Entries = append(standings.Entries, StandingEntry{
			EntryID:     entry.ID,
			DisplayName: entry.DisplayName,
			InOvers:     entry.InOvers,
			TotalPayout: total,
			FRLBonus:    entryFRLBonus,
			Picks:       standingPicks,
		})
	}

	sort.Slice(standings.Entries, func(i, j int) bool {
		if standings.Entries[i].TotalPayout == standings.Entries[j].TotalPayout {
			return standings.Entries[i].DisplayName < standings.Entries[j].DisplayName
		}

		return standings.Entries[i].TotalPayout > standings.Entries[j].TotalPayout
	})

	rank := 1
	for i := range standings.Entries {
		if i > 0 && standings.Entries[i].TotalPayout < standings.Entries[i-1].TotalPayout {
			rank = i + 1
		}
		standings.Entries[i].Rank = rank
	}

	return standings, nil
}

type entryPick struct {
	GroupName  string
	GolferName string
}

func collectPicks(raw map[string]any) ([]entryPick, []string) {
	picks := make([]entryPick, 0, len(raw))
	golfers := make([]string, 0, len(raw))

	for groupName, rawPick := range raw {
		golferName := strings.TrimSpace(extractPickName(rawPick))
		if golferName == "" {
			continue
		}

		picks = append(picks, entryPick{
			GroupName:  groupName,
			GolferName: golferName,
		})
		golfers = append(golfers, golferName)
	}

	return picks, golfers
}

func extractPickName(raw any) string {
	switch typed := raw.(type) {
	case string:
		return typed
	case map[string]any:
		for _, key := range []string{"name", "player", "label", "value"} {
			if candidate, ok := typed[key].(string); ok {
				return candidate
			}
		}
	}

	return ""
}

func payoutMultiplier(groupName string) int64 {
	normalized := normalizeName(groupName)
	switch normalized {
	case "mutt":
		return 2
	case "oldmutt":
		return 3
	default:
		return 1
	}
}

func projectedPayouts(results []db.GolferResult, payouts map[int]int64) map[string]int64 {
	type group struct {
		rank    int
		results []db.GolferResult
	}

	groups := make(map[int][]db.GolferResult)
	ranks := []int{}
	for _, result := range results {
		rank, ok := parseRank(result.Position)
		if !ok {
			continue
		}

		if _, exists := groups[rank]; !exists {
			ranks = append(ranks, rank)
		}
		groups[rank] = append(groups[rank], result)
	}

	sort.Ints(ranks)

	projected := make(map[string]int64, len(results))
	for _, rank := range ranks {
		groupResults := groups[rank]
		size := len(groupResults)
		if size == 0 {
			continue
		}

		total := int64(0)
		for position := rank; position < rank+size; position++ {
			total += payouts[position]
		}

		share := int64(math.Round(float64(total) / float64(size)))
		for _, result := range groupResults {
			projected[normalizeName(result.GolferName)] = share
		}
	}

	return projected
}

func parsePoolPayouts(raw map[string]any) (map[int]int64, error) {
	parsed := make(map[int]int64, len(raw))
	for key, value := range raw {
		position, err := strconv.Atoi(strings.TrimSpace(key))
		if err != nil {
			return nil, fmt.Errorf("pool payout key %q is not a valid position", key)
		}

		amount, err := toInt64(value)
		if err != nil {
			return nil, fmt.Errorf("pool payout for position %d: %w", position, err)
		}

		parsed[position] = amount
	}

	return parsed, nil
}

func toInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case int:
		return int64(typed), nil
	case int64:
		return typed, nil
	case float64:
		return int64(math.Round(typed)), nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", value)
	}
}

func parseRank(position string) (int, bool) {
	text := strings.ToUpper(strings.TrimSpace(position))
	if text == "" {
		return 0, false
	}

	if strings.HasPrefix(text, "T") {
		text = text[1:]
	}

	digits := strings.Builder{}
	for _, r := range text {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
			continue
		}
		break
	}

	if digits.Len() == 0 {
		return 0, false
	}

	rank, err := strconv.Atoi(digits.String())
	if err != nil {
		return 0, false
	}

	return rank, rank > 0
}

func normalizeName(value string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

func resolveGolferKey(value string, aliases map[string]string) string {
	normalized := normalizeName(value)
	if resolved, ok := aliases[normalized]; ok {
		return resolved
	}

	return normalized
}

func buildNameAliases(results []db.GolferResult) map[string]string {
	aliases := make(map[string]string, len(results)*3)
	lastNameCounts := make(map[string]int)

	for _, result := range results {
		parts := splitNameParts(result.GolferName)
		if len(parts) == 0 {
			continue
		}

		lastName := normalizeName(parts[len(parts)-1])
		if lastName != "" {
			lastNameCounts[lastName]++
		}
	}

	for _, result := range results {
		fullKey := normalizeName(result.GolferName)
		if fullKey == "" {
			continue
		}

		aliases[fullKey] = fullKey

		parts := splitNameParts(result.GolferName)
		if len(parts) == 0 {
			continue
		}

		lastName := normalizeName(parts[len(parts)-1])
		if lastName != "" && lastNameCounts[lastName] == 1 {
			aliases[lastName] = fullKey
		}

		if len(parts) >= 2 {
			firstInitialLast := normalizeName(string(parts[0][0]) + parts[len(parts)-1])
			if firstInitialLast != "" {
				aliases[firstInitialLast] = fullKey
			}
		}
	}

	return aliases
}

func splitNameParts(value string) []string {
	return strings.Fields(strings.TrimSpace(value))
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
