package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// wc26Disabled lets an operator turn off the community live feed (e.g. if the
// volunteer server is down) without touching the authoritative results path.
// Set WC26_LIVE=off (or 0/false) to disable.
func wc26Disabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("WC26_LIVE"))) {
	case "off", "0", "false", "no":
		return true
	}
	return false
}

// worldcup26.ir is a free, community-run REST API (https://github.com/rezarahiminia/worldcup2026)
// that publishes *in-progress* match scores — the one thing the openfootball
// feed can't give us. We use it ONLY as a live-score source: it sets a match's
// running score and status="live" while it's being played, but never finalizes
// (never sets finalizedAt). Final results — including knockout extra-time and
// penalties, which this feed doesn't break out — stay the job of the
// openfootball / API-Football path, so the league table is untouched until a
// match is properly finalized there.
//
// To respect a volunteer-run server, this is only ever called from the live-
// window cron (gated by anyMatchLive), so it's hit at most once every few
// minutes and only while a match is actually in progress.
const wc26URL = "https://worldcup26.ir/get/games"

type wc26Game struct {
	ID          string `json:"id"`
	HomeScore   string `json:"home_score"`
	AwayScore   string `json:"away_score"`
	Group       string `json:"group"`
	Finished    string `json:"finished"`
	TimeElapsed string `json:"time_elapsed"`
	Type        string `json:"type"`
	HomeNameEn  string `json:"home_team_name_en"`
	AwayNameEn  string `json:"away_team_name_en"`
}

// inProgress reports whether the feed considers this game live (kicked off,
// not yet finished). Finished games are left to the authoritative finalizer.
func (g wc26Game) inProgress() bool {
	if strings.EqualFold(g.Finished, "TRUE") {
		return false
	}
	te := strings.ToLower(strings.TrimSpace(g.TimeElapsed))
	return te != "" && te != "notstarted"
}

// worldcup26LiveSync pulls current scores and marks in-progress matches live.
// Idempotent: a record is saved only when its score or status actually changes.
func worldcup26LiveSync(ctx context.Context, app core.App) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wc26URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "wm-tips/1.0")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return fmt.Errorf("worldcup26 fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("worldcup26: status %d", resp.StatusCode)
	}
	var doc struct {
		Games []wc26Game `json:"games"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return err
	}

	// Only unfinalized matches are candidates — we never override an
	// authoritative result.
	matches, err := app.FindRecordsByFilter("matches", "finalizedAt = ''", "", 0, 0)
	if err != nil {
		return err
	}
	teamName := map[string]string{} // teamId -> canonName
	teams, err := app.FindRecordsByFilter("teams", "id != ''", "", 0, 0)
	if err != nil {
		return err
	}
	for _, t := range teams {
		teamName[t.Id] = canonName(t.GetString("name"))
	}
	byPair := map[string]*core.Record{} // group stage: canonHome|canonAway
	byNum := map[int]*core.Record{}     // knockout: FIFA match number
	for _, m := range matches {
		if n := m.GetInt("num"); n > 0 {
			byNum[n] = m
		}
		h, a := teamName[m.GetString("homeTeam")], teamName[m.GetString("awayTeam")]
		if h != "" && a != "" {
			byPair[h+"|"+a] = m
		}
	}

	updated := 0
	for _, g := range doc.Games {
		if !g.inProgress() {
			continue
		}
		// Match by FIFA number (knockout rows carry num 73+), then fall back to
		// the team-name pair — which covers every group game and any knockout
		// row we didn't seed a num for (e.g. the final / third-place match).
		var rec *core.Record
		if n, err := strconv.Atoi(strings.TrimSpace(g.ID)); err == nil {
			rec = byNum[n]
		}
		if rec == nil {
			rec = byPair[canonName(g.HomeNameEn)+"|"+canonName(g.AwayNameEn)]
		}
		if rec == nil {
			continue
		}
		hs, errH := strconv.Atoi(strings.TrimSpace(g.HomeScore))
		as, errA := strconv.Atoi(strings.TrimSpace(g.AwayScore))
		if errH != nil || errA != nil {
			continue
		}
		changed := false
		if rec.GetInt("ftHome") != hs {
			rec.Set("ftHome", hs)
			changed = true
		}
		if rec.GetInt("ftAway") != as {
			rec.Set("ftAway", as)
			changed = true
		}
		if rec.GetString("status") != "live" {
			rec.Set("status", "live")
			changed = true
		}
		if changed && app.Save(rec) == nil {
			updated++
		}
	}
	if updated > 0 {
		log.Printf("[sync] worldcup26 live: updated=%d", updated)
	}
	return nil
}
