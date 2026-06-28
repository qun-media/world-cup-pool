package sync

import (
	"log"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"

	"github.com/oyvhov/world-cup-pool/internal/bracket"
	"github.com/oyvhov/world-cup-pool/internal/standings"
)

// ResolveBracket fills knockout matches' homeTeam/awayTeam from their
// placeholder labels once the referenced results are known. This is what makes
// a knockout Tip become available (Phase 3): a Tip opens as soon as both teams
// of a matchup are resolved.
//
// Resolvable labels:
//   - "1A".."2L"      group winner / runner-up (once that group is complete)
//   - "3A/B/C/D/F"     a best-third slot (see note)
//   - "W73" / "L101"   winner / loser of a finished knockout match
//
// Group order (and so who is 1A/2A/3A) follows the full FIFA tiebreaker chain,
// including head-to-head, via internal/standings — shared with Forecast scoring.
//
// NOTE: once all 8 best thirds are known, the best-third -> R32 slot mapping
// uses FIFA's official Annex C combination table (internal/bracket). While the
// group stage is still incomplete it falls back to a deterministic greedy fill,
// which only runs when the bracket can't be resolved yet anyway.
func ResolveBracket(app core.App) error {
	matches, err := app.FindRecordsByFilter("matches", "id != ''", "num", 0, 0)
	if err != nil {
		return err
	}

	byNum := map[int]*core.Record{}
	for _, m := range matches {
		if n := m.GetInt("num"); n > 0 {
			byNum[n] = m
		}
	}

	first, second, thirds, thirdTeam, groupStageComplete := groupStandings(matches)

	// Resolve the 8 R32 third-slots. A best-third slot is only knowable once the
	// ENTIRE group stage is finished: the 8th-vs-9th third cut depends on all 12
	// groups, so any earlier assignment (from a partial set of thirds) would be a
	// guess that the official table later contradicts. We therefore resolve the
	// third-slots exclusively from FIFA's official Annex C table, and only once
	// every group is complete — at which point the lookup is guaranteed to hit.
	thirdByNum := map[int]string{}
	if groupStageComplete {
		quals := make([]string, 0, len(thirds))
		for _, st := range thirds {
			quals = append(quals, st.group)
		}
		if tbl, ok := bracket.Lookup(quals); ok {
			for _, m := range matches {
				if m.GetString("stage") != "R32" {
					continue
				}
				home, away := m.GetString("homeLabel"), m.GetString("awayLabel")
				if w, ok := bracket.WinnerLetter(home, away); ok {
					thirdByNum[m.GetInt("num")] = thirdTeam[tbl[w]]
				}
			}
		}
	}

	resolve := func(label string, num int) string {
		if label == "" {
			return ""
		}
		switch label[0] {
		case '1':
			return first[label[1:]]
		case '2':
			return second[label[1:]]
		case '3':
			return thirdByNum[num]
		case 'W', 'L':
			n, err := strconv.Atoi(label[1:])
			if err != nil {
				return ""
			}
			src, ok := byNum[n]
			if !ok || src.GetString("finalizedAt") == "" {
				return ""
			}
			adv := src.GetString("advancer")
			if label[0] == 'W' {
				return adv
			}
			// loser = the side that is not the advancer
			h, a := src.GetString("homeTeam"), src.GetString("awayTeam")
			if adv == h {
				return a
			}
			if adv == a {
				return h
			}
			return ""
		}
		return ""
	}

	// isThirdSlot reports whether a side's label is a best-third placeholder
	// (e.g. "3A/B/C/D/F"). Such a side is resolved authoritatively from the
	// official table once the group stage is complete, so we OVERWRITE any value
	// a previous (premature) run may have written — every other label resolves
	// to a stable team and so is only filled when still empty.
	isThirdSlot := func(label string) bool {
		return strings.HasPrefix(label, "3") && strings.Contains(label, "/")
	}
	setSide := func(m *core.Record, field, label string, num int) bool {
		if isThirdSlot(label) {
			if !groupStageComplete {
				return false
			}
			id := thirdByNum[num]
			if id == "" || m.GetString(field) == id {
				return false
			}
			m.Set(field, id)
			return true
		}
		if m.GetString(field) != "" {
			return false
		}
		id := resolve(label, num)
		if id == "" {
			return false
		}
		m.Set(field, id)
		return true
	}

	for _, m := range matches {
		if m.GetString("stage") == "group" {
			continue
		}
		num := m.GetInt("num")
		changed := setSide(m, "homeTeam", m.GetString("homeLabel"), num)
		if setSide(m, "awayTeam", m.GetString("awayLabel"), num) {
			changed = true
		}
		if changed {
			if err := app.Save(m); err != nil {
				return err
			}
		}
	}
	return nil
}

type standing struct {
	group string
	team  string
}

// groupStandings computes, from finished group matches only, the 1st/2nd team
// id per group letter (only when that group's 6 matches are all finished) plus
// the globally ranked list of the best third-placed teams (top 8) and a
// group-letter -> third-placed team id map (thirdTeam, used by the best-third
// allocation in ResolveBracket). It delegates the FIFA tiebreaker order
// (including head-to-head) to internal/standings so bracket resolution and
// Forecast scoring always agree, and logs any group that needed a non-official
// tiebreak so an admin can verify/override.
func groupStandings(matches []*core.Record) (first, second map[string]string, thirds []standing, thirdTeam map[string]string, complete bool) {
	first = map[string]string{}
	second = map[string]string{}
	thirdTeam = map[string]string{}

	order, ranked, ambiguous := standings.GroupTables(standings.FromRecords(matches))
	// GroupTables only includes fully-played groups, so all 12 present means the
	// entire group stage is finished and the best-third cut is final.
	complete = len(order) == 12
	for g, ids := range order {
		if len(ids) >= 2 {
			first[g] = ids[0]
			second[g] = ids[1]
		}
	}
	for _, r := range ranked {
		thirdTeam[r.Group] = r.TeamID
		thirds = append(thirds, standing{group: r.Group, team: r.TeamID})
	}
	if len(thirds) > 8 {
		thirds = thirds[:8]
	}
	if len(ambiguous) > 0 {
		log.Printf("[sync] group ranking needed a non-official tiebreak (fair play / lots) for %v — verify and override if needed", ambiguous)
	}
	return first, second, thirds, thirdTeam, complete
}
