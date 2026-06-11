// Package forecast backs the one-time pre-tournament prediction: full group
// standings (1..4 per group), the 8 manually-chosen best-third R32 slots, and
// the knockout bracket winners. It is editable until the tournament's first
// kickoff (global lock) and validated server-side.
package forecast

import (
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"

	"github.com/oyvhov/world-cup-pool/internal/bracket"
	"github.com/oyvhov/world-cup-pool/internal/clock"
	"github.com/oyvhov/world-cup-pool/internal/scoring"
)

// tournamentStart returns the earliest match kickoff (the global Forecast
// deadline).
func tournamentStart(app core.App) (time.Time, error) {
	ms, err := app.FindRecordsByFilter("matches", "id != ''", "kickoff", 1, 0)
	if err != nil || len(ms) == 0 {
		return time.Time{}, err
	}
	return ms[0].GetDateTime("kickoff").Time(), nil
}

func locked(app core.App) bool {
	ts, err := tournamentStart(app)
	if err != nil {
		return false
	}
	return isLocked(clock.Now(app), ts)
}

func isLocked(now, start time.Time) bool {
	return !now.Before(start)
}

// Locked reports whether the global Forecast deadline (first kickoff) has
// passed. Exported so leagues can decide when to surface the admin
// "unlock forecast" control.
func Locked(app core.App) bool { return locked(app) }

// userUnlocked reports whether any league admin has flagged this user's
// membership as forecastUnlocked, temporarily re-opening their Forecast. The
// Forecast is a single per-user record, so an unlock granted in any league
// applies everywhere.
func userUnlocked(app core.App, userID string) bool {
	if userID == "" {
		return false
	}
	_, err := app.FindFirstRecordByFilter("league_members",
		"user = {:u} && forecastUnlocked = true", map[string]any{"u": userID})
	return err == nil
}

// groupTeams returns letter -> set(teamId) from tournament_groups.
func groupTeams(app core.App) (map[string]map[string]bool, error) {
	gs, err := app.FindRecordsByFilter("tournament_groups", "id != ''", "letter", 0, 0)
	if err != nil {
		return nil, err
	}
	out := map[string]map[string]bool{}
	for _, g := range gs {
		set := map[string]bool{}
		for _, id := range g.GetStringSlice("teams") {
			set[id] = true
		}
		out[g.GetString("letter")] = set
	}
	return out, nil
}

// validate enforces the lock and that group orderings only contain that
// group's own teams without duplicates. Partial forecasts are allowed (the
// user fills it in over multiple sessions); only clearly invalid data is
// rejected.
// bypass lets the dev bot generator insert a complete Forecast regardless of
// the lock. Never set in production (dev-only path).
var bypass atomic.Bool

// SetBypass toggles the dev-only validation bypass.
func SetBypass(b bool) { bypass.Store(b) }

func validate(app core.App, rec *core.Record) error {
	if bypass.Load() {
		return nil
	}
	if locked(app) && !userUnlocked(app, rec.GetString("user")) {
		return apis.NewBadRequestError("the tournament has started — the Forecast is locked", nil)
	}
	groups, err := groupTeams(app)
	if err != nil {
		return err
	}
	var order map[string][]string
	if err := rec.UnmarshalJSONField("groupOrder", &order); err != nil {
		return nil // empty/!set yet — allow
	}
	for letter, ids := range order {
		members := groups[letter]
		if members == nil {
			return apis.NewBadRequestError("unknown group "+letter, nil)
		}
		seen := map[string]bool{}
		for _, id := range ids {
			if id == "" {
				continue
			}
			if !members[id] {
				return apis.NewBadRequestError("a team in group "+letter+" is not in that group", nil)
			}
			if seen[id] {
				return apis.NewBadRequestError("duplicate team in group "+letter, nil)
			}
			seen[id] = true
		}
	}
	return nil
}

// ThirdSlot is an R32 match whose away side is a best-third placeholder.
type ThirdSlot struct {
	MatchNum int      `json:"matchNum"`
	Winner   string   `json:"winner"`  // group-winner letter this slot pairs with
	Allowed  []string `json:"allowed"` // group letters eligible (fallback only)
}

// canViewForecast reports whether user 'viewer' may see 'target's forecast.
// Access is granted when they share at least one league that does not have
// hideForecast enabled. If every shared league has hideForecast=true (or they
// share no league at all), access is denied.
func canViewForecast(app core.App, viewer, target string) bool {
	mine, err := app.FindRecordsByFilter("league_members",
		"user = {:u}", "", 0, 0, map[string]any{"u": viewer})
	if err != nil {
		return false
	}
	for _, m := range mine {
		leagueID := m.GetString("league")
		if _, err := app.FindFirstRecordByFilter("league_members",
			"league = {:l} && user = {:u}",
			map[string]any{"l": leagueID, "u": target}); err != nil {
			continue // target not in this league
		}
		league, err := app.FindRecordById("leagues", leagueID)
		if err != nil {
			continue
		}
		if !league.GetBool("hideForecast") {
			return true // found a shared league where forecasts are visible
		}
	}
	return false
}

// Register wires the Forecast validation hooks and the structure endpoint.
func Register(app core.App, se *core.ServeEvent) {
	app.OnRecordCreate("forecasts").BindFunc(func(e *core.RecordEvent) error {
		if err := validate(e.App, e.Record); err != nil {
			return err
		}
		return e.Next()
	})
	app.OnRecordUpdate("forecasts").BindFunc(func(e *core.RecordEvent) error {
		if err := validate(e.App, e.Record); err != nil {
			return err
		}
		return e.Next()
	})

	// When a Forecast is saved after the global lock (only possible for an
	// admin-unlocked member), recompute so the leaderboard reflects any
	// changed picks for matches that may already be scored. Pre-lock autosaves
	// are skipped — results don't exist yet, so there's nothing to refresh.
	recomputeIfLocked := func(e *core.RecordEvent) error {
		if locked(e.App) {
			if err := scoring.Recompute(e.App); err != nil {
				log.Printf("[forecast] recompute after unlocked save: %v", err)
			}
		}
		return e.Next()
	}
	app.OnRecordAfterUpdateSuccess("forecasts").BindFunc(recomputeIfLocked)
	app.OnRecordAfterCreateSuccess("forecasts").BindFunc(recomputeIfLocked)

	// GET /api/forecast/of/{userId} — a friend's Forecast. Visible to anyone
	// who shares a League with them (no lock gate: in a friends group you
	// want to see picks right away). Not registered on the forecasts table,
	// which stays own-only.
	se.Router.GET("/api/forecast/of/{userId}", func(e *core.RequestEvent) error {
		uid := e.Request.PathValue("userId")
		if uid != e.Auth.Id && !canViewForecast(app, e.Auth.Id, uid) {
			return apis.NewForbiddenError("forecast not available", nil)
		}
		u, err := app.FindRecordById("users", uid)
		if err != nil {
			return apis.NewNotFoundError("user not found", nil)
		}
		out := map[string]any{"userId": uid, "name": u.GetString("name")}
		fc, err := app.FindFirstRecordByFilter("forecasts",
			"user = {:u}", map[string]any{"u": uid})
		if err != nil {
			out["forecast"] = nil
			return e.JSON(http.StatusOK, out)
		}
		var order, bracket map[string]any
		var thirds map[string]any
		parse := func(field string, dst any) {
			if raw := fc.GetString(field); raw == "" || raw == "null" {
				return
			}
			if err := fc.UnmarshalJSONField(field, dst); err != nil {
				log.Printf("[forecast] %s: bad %s JSON: %v", fc.Id, field, err)
			}
		}
		parse("groupOrder", &order)
		parse("thirdQualifiers", &thirds)
		parse("bracket", &bracket)
		out["forecast"] = map[string]any{
			"groupOrder":      order,
			"thirdQualifiers": thirds,
			"bracket":         bracket,
		}
		return e.JSON(http.StatusOK, out)
	}).Bind(apis.RequireAuth())

	// GET /api/forecast/structure — everything the builder needs: groups with
	// their teams, the knockout match skeleton with placeholder labels, the
	// best-third slots with their eligible groups, and the lock state.
	se.Router.GET("/api/forecast/structure", func(e *core.RequestEvent) error {
		groups, err := app.FindRecordsByFilter("tournament_groups", "id != ''", "letter", 0, 0)
		if err != nil {
			return err
		}
		gOut := make([]map[string]any, 0, len(groups))
		for _, g := range groups {
			gOut = append(gOut, map[string]any{
				"letter": g.GetString("letter"),
				"teams":  g.GetStringSlice("teams"),
			})
		}

		ko, err := app.FindRecordsByFilter("matches", "stage != 'group'", "num", 0, 0)
		if err != nil {
			return err
		}
		kOut := make([]map[string]any, 0, len(ko))
		thirds := make([]ThirdSlot, 0, 8)
		for _, mt := range ko {
			home := mt.GetString("homeLabel")
			away := mt.GetString("awayLabel")
			num := mt.GetInt("num")
			kOut = append(kOut, map[string]any{
				"num":       num,
				"stage":     mt.GetString("stage"),
				"round":     mt.GetString("roundLabel"),
				"homeLabel": home,
				"awayLabel": away,
			})
			for _, lbl := range []string{home, away} {
				if strings.HasPrefix(lbl, "3") && strings.Contains(lbl, "/") {
					w, _ := bracket.WinnerLetter(home, away)
					thirds = append(thirds, ThirdSlot{
						MatchNum: num,
						Winner:   w,
						Allowed:  strings.Split(strings.TrimPrefix(lbl, "3"), "/"),
					})
				}
			}
		}

		ts, _ := tournamentStart(app)
		return e.JSON(http.StatusOK, map[string]any{
			"groups":          gOut,
			"knockout":        kOut,
			"thirdSlots":      thirds,
			"thirdTable":      bracket.Table(),
			"tournamentStart": ts,
			"locked":          locked(app),
			"unlocked":        locked(app) && userUnlocked(app, e.Auth.Id),
		})
	}).Bind(apis.RequireAuth())
}
