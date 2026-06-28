package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/oyvhov/world-cup-pool/internal/seed"
)

// openfootball is the free live-results source: the same project we seed
// from publishes scores into 2026/worldcup.json during the tournament.
// Matches map 1:1 to our rows by the shared deterministic ExtID (no team
// name aliasing), and its `score.et` is already the cumulative after-120
// score — exactly our model.
const ofLiveURL = "https://raw.githubusercontent.com/openfootball/worldcup.json/master/2026/worldcup.json"

type ofScore struct {
	FT []int `json:"ft"`
	ET []int `json:"et"`
	P  []int `json:"p"`
}
type ofLiveMatch struct {
	Round string   `json:"round"`
	Num   int      `json:"num"`
	Team1 string   `json:"team1"`
	Team2 string   `json:"team2"`
	Group string   `json:"group"`
	Score *ofScore `json:"score"`
}

func pi(v int) *int { return &v }

// fetchOpenfootballMatches pulls and decodes openfootball's worldcup.json.
func fetchOpenfootballMatches(ctx context.Context) ([]ofLiveMatch, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ofLiveURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wm-tips/1.0")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("openfootball fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openfootball: status %d", resp.StatusCode)
	}
	var doc struct {
		Matches []ofLiveMatch `json:"matches"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	return doc.Matches, nil
}

// openfootballSync pulls openfootball's live JSON and applies any results.
// Idempotent: a record is only saved when something actually changed.
func openfootballSync(ctx context.Context, app core.App) error {
	ofMatches, err := fetchOpenfootballMatches(ctx)
	if err != nil {
		return err
	}

	byExt := map[string]*core.Record{}
	recs, err := app.FindRecordsByFilter("matches", "id != ''", "", 0, 0)
	if err != nil {
		return err
	}
	for _, r := range recs {
		byExt[r.GetString("extId")] = r
	}

	updated := 0
	for _, m := range ofMatches {
		if m.Score == nil || len(m.Score.FT) != 2 {
			continue // not played yet
		}
		rec := byExt[seed.ExtID(m.Round, m.Num, m.Group, m.Team1, m.Team2)]
		if rec == nil {
			continue
		}
		ftH, ftA := m.Score.FT[0], m.Score.FT[1]
		var etH, etA, penH, penA *int
		if len(m.Score.ET) == 2 { // cumulative after-120
			etH, etA = pi(m.Score.ET[0]), pi(m.Score.ET[1])
		}
		if len(m.Score.P) == 2 {
			penH, penA = pi(m.Score.P[0]), pi(m.Score.P[1])
		}
		// Skip if nothing changed (avoids needless recompute storms).
		if rec.GetString("status") == "finished" &&
			rec.GetInt("ftHome") == ftH && rec.GetInt("ftAway") == ftA &&
			rec.GetInt("penHome") == ip(penH) && rec.GetInt("penAway") == ip(penA) &&
			rec.GetInt("etHome") == ip(etH) && rec.GetInt("etAway") == ip(etA) {
			continue
		}
		applyResult(rec, "finished", pi(ftH), pi(ftA), etH, etA, penH, penA)
		if app.Save(rec) == nil {
			updated++
		}
	}
	if err := ResolveBracket(app); err != nil {
		return err
	}
	// Defence-in-depth: cross-check our computed knockout matchups against
	// openfootball's published teams and correct any disagreement.
	if err := reconcileKnockoutTeams(app, ofMatches); err != nil {
		return err
	}
	return nil
}

// verifyKnockoutFromOpenfootball fetches openfootball and reconciles our
// knockout matchups against it. Used by results sources other than openfootball
// (e.g. API-Football) so the cross-check runs no matter where scores come from.
func verifyKnockoutFromOpenfootball(ctx context.Context, app core.App) error {
	ofMatches, err := fetchOpenfootballMatches(ctx)
	if err != nil {
		return err
	}
	return reconcileKnockoutTeams(app, ofMatches)
}

// reconcileKnockoutTeams treats openfootball as the authoritative oracle for
// who plays whom in the knockout stage. openfootball fills each knockout match's
// team1/team2 with real team names once the feeding round is decided (keyed by
// the same `num` we use), so once both sides resolve to teams we recognise we
// make our record agree — logging loudly on any correction. We only touch
// matches that haven't been played yet: a finished match's teams are already
// settled, and rewriting them would corrupt its result. This is a safety net
// over our own standings-based ResolveBracket, not a replacement for it (early
// rounds still resolve instantly from results before openfootball updates).
// koCorrection decides whether one openfootball knockout match should drive a
// correction, and the team ids it should resolve to. It returns ok=false when
// either side is still an openfootball placeholder (W74, 1A, 3A/B/… — not a name
// we recognise) or when the match is already played (teams settled).
func koCorrection(ofTeam1, ofTeam2 string, nameToID map[string]string, finished bool) (homeID, awayID string, ok bool) {
	h, okH := nameToID[ofTeam1]
	a, okA := nameToID[ofTeam2]
	if !okH || !okA || finished {
		return "", "", false
	}
	return h, a, true
}

func reconcileKnockoutTeams(app core.App, ofMatches []ofLiveMatch) error {
	teams, err := app.FindRecordsByFilter("teams", "id != ''", "", 0, 0)
	if err != nil {
		return err
	}
	nameToID := make(map[string]string, len(teams))
	for _, t := range teams {
		nameToID[t.GetString("name")] = t.Id
	}

	byNum := map[int]*core.Record{}
	recs, err := app.FindRecordsByFilter("matches", "stage != 'group'", "", 0, 0)
	if err != nil {
		return err
	}
	for _, r := range recs {
		if n := r.GetInt("num"); n > 0 {
			byNum[n] = r
		}
	}

	for _, m := range ofMatches {
		rec := byNum[m.Num]
		if rec == nil || rec.GetString("stage") == "group" {
			continue
		}
		finished := rec.GetString("finalizedAt") != "" || rec.GetString("status") == "finished"
		homeID, awayID, ok := koCorrection(m.Team1, m.Team2, nameToID, finished)
		if !ok {
			continue
		}
		changed := false
		if rec.GetString("homeTeam") != homeID {
			log.Printf("[sync] knockout match %d home corrected to openfootball: %q -> %q",
				m.Num, rec.GetString("homeTeam"), m.Team1)
			rec.Set("homeTeam", homeID)
			changed = true
		}
		if rec.GetString("awayTeam") != awayID {
			log.Printf("[sync] knockout match %d away corrected to openfootball: %q -> %q",
				m.Num, rec.GetString("awayTeam"), m.Team2)
			rec.Set("awayTeam", awayID)
			changed = true
		}
		if changed {
			if err := app.Save(rec); err != nil {
				return err
			}
		}
	}
	return nil
}
