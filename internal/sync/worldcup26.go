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

// The worldcup26 community API (https://github.com/rezarahiminia/worldcup2026)
// is a free, volunteer-run REST service that publishes *in-progress* match
// scores — the one thing the openfootball feed can't give us. We use it ONLY as
// a live-score source: it sets a match's running score and status="live" while
// it's being played, but never finalizes (never sets finalizedAt). Final
// results — including knockout extra-time and penalties, which this feed doesn't
// break out — stay the job of the openfootball / API-Football path, so the
// league table is untouched until a match is properly finalized there.
//
// These are volunteer servers that do go down (HTTP 500/502), so we try the
// mirrors below in order: they run the same software with an identical
// /get/games schema, so a fallback is a drop-in. The documented rate limit is
// generous (500 requests / minute) and we stay far under it: this is only ever
// called from the live-window cron (gated by anyMatchLive) — at most once a
// minute and only while a match is actually in progress.
var wc26URLs = []string{
	"https://worldcup26.ir/get/games",     // primary
	"https://wc2026.moothz.win/get/games", // backup mirror (same software)
}

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

// unfinished reports whether the feed has NOT yet marked this game finished.
// We deliberately ignore the feed's `time_elapsed` clock — it lags and is often
// "notstarted" even after a goal — and instead decide "has it kicked off?" from
// our own match kickoff time (see worldcup26LiveSync). That picks up scores
// sooner without trusting an unreliable field. Finished games are always left
// to the authoritative finalizer.
func (g wc26Game) unfinished() bool {
	return !strings.EqualFold(g.Finished, "TRUE")
}

// fetchWc26Games returns the games list from the first mirror that answers with
// a usable payload, trying wc26URLs in order. The returned string is the URL the
// data came from (for logging). An error means every mirror failed this tick.
func fetchWc26Games(ctx context.Context) ([]wc26Game, string, error) {
	return fetchWc26GamesFrom(ctx, wc26URLs)
}

func fetchWc26GamesFrom(ctx context.Context, urls []string) ([]wc26Game, string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	var errs []string
	for _, url := range urls {
		games, err := fetchWc26One(ctx, client, url)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", url, err))
			continue
		}
		return games, url, nil
	}
	return nil, "", fmt.Errorf("all worldcup26 mirrors failed: %s", strings.Join(errs, "; "))
}

// fetchWc26One fetches and decodes a single mirror. An empty games list counts
// as a failure so we fall through to the next mirror — these feeds always carry
// all 104 fixtures, so empty means a broken/stale deployment (or a mirror whose
// schema we don't recognise).
func fetchWc26One(ctx context.Context, client *http.Client, url string) ([]wc26Game, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wm-tips/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var doc struct {
		Games []wc26Game `json:"games"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	if len(doc.Games) == 0 {
		return nil, fmt.Errorf("empty games list")
	}
	return doc.Games, nil
}

// worldcup26LiveSync pulls current scores and marks in-progress matches live.
// Idempotent: a record is saved only when its score or status actually changes.
func worldcup26LiveSync(ctx context.Context, app core.App) error {
	games, src, err := fetchWc26Games(ctx)
	if err != nil {
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

	now := time.Now().UTC()
	updated := 0
	for _, g := range games {
		if !g.unfinished() {
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
		// Has it actually kicked off? We trust our own kickoff time rather than
		// the feed's clock, so a not-yet-started match never gets marked live
		// (and stamped with a stray 0–0) just because the feed lists it.
		kickoff := rec.GetDateTime("kickoff").Time()
		if kickoff.IsZero() || now.Before(kickoff) {
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
		log.Printf("[sync] worldcup26 live: updated=%d src=%s", updated, src)
	}
	return nil
}
