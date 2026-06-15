package tips

import (
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"

	"github.com/oyvhov/world-cup-pool/internal/scoring"
)

func liveMatchRecord(stage, status string, kickoff time.Time, finalized bool) *core.Record {
	c := core.NewBaseCollection("matches")
	c.Fields.Add(&core.TextField{Name: "stage", Max: 16})
	c.Fields.Add(&core.TextField{Name: "status", Max: 16})
	c.Fields.Add(&core.DateField{Name: "kickoff"})
	c.Fields.Add(&core.DateField{Name: "finalizedAt"})
	r := core.NewRecord(c)
	r.Set("stage", stage)
	r.Set("status", status)
	r.Set("kickoff", kickoff)
	if finalized {
		r.Set("finalizedAt", kickoff.Add(2*time.Hour))
	}
	return r
}

func TestIsLiveNow(t *testing.T) {
	ko := time.Date(2026, time.June, 11, 19, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		rec  *core.Record
		now  time.Time
		want bool
	}{
		{"before kickoff", liveMatchRecord("group", "scheduled", ko, false), ko.Add(-time.Minute), false},
		{"just after kickoff", liveMatchRecord("group", "scheduled", ko, false), ko.Add(time.Minute), true},
		{"inside group window", liveMatchRecord("group", "scheduled", ko, false), ko.Add(149 * time.Minute), true},
		{"past group window", liveMatchRecord("group", "scheduled", ko, false), ko.Add(151 * time.Minute), false},
		{"KO uses longer window", liveMatchRecord("R32", "scheduled", ko, false), ko.Add(200 * time.Minute), true},
		{"status live overrides window", liveMatchRecord("group", "live", ko, false), ko.Add(300 * time.Minute), true},
		{"finalized is never live", liveMatchRecord("group", "live", ko, true), ko.Add(time.Minute), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isLiveNow(tc.rec, tc.now); got != tc.want {
				t.Fatalf("isLiveNow() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestScoreTipPointsProjectsCurrentScore(t *testing.T) {
	mc := core.NewBaseCollection("matches")
	for _, f := range []string{"stage", "advancer"} {
		mc.Fields.Add(&core.TextField{Name: f, Max: 32})
	}
	for _, f := range []string{"ftHome", "ftAway", "etHome", "etAway"} {
		mc.Fields.Add(&core.NumberField{Name: f, OnlyInt: true})
	}
	match := core.NewRecord(mc)
	match.Set("stage", "group")
	match.Set("ftHome", 2)
	match.Set("ftAway", 1)

	tc := core.NewBaseCollection("tips")
	for _, f := range []string{"penWinner", "advancer"} {
		tc.Fields.Add(&core.TextField{Name: f, Max: 32})
	}
	for _, f := range []string{"ftHome", "ftAway", "etHome", "etAway"} {
		tc.Fields.Add(&core.NumberField{Name: f, OnlyInt: true})
	}
	tip := core.NewRecord(tc)
	tip.Set("ftHome", 2)
	tip.Set("ftAway", 1)

	var cfg scoring.Config
	cfg.Match.Tendency = 3
	cfg.Match.Exact = 1
	cfg.Match.TotalGoals = 1
	cfg.Match.GoalDiff = 1

	if got := scoring.ScoreTipPoints(cfg, match, tip); got != 6 {
		t.Fatalf("ScoreTipPoints (exact group tip) = %d, want 6", got)
	}

	// Same tip, different live score (2-1 vs 3-2): tendency + goal-diff match,
	// exact + total goals do not → 3 + 1 = 4.
	match.Set("ftHome", 3)
	match.Set("ftAway", 2)
	if got := scoring.ScoreTipPoints(cfg, match, tip); got != 4 {
		t.Fatalf("ScoreTipPoints (tendency+diff) = %d, want 4", got)
	}
}

func TestIsLockedAtKickoffBoundary(t *testing.T) {
	kickoff := time.Date(2026, time.June, 11, 19, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "before kickoff remains editable", now: kickoff.Add(-time.Nanosecond), want: false},
		{name: "exact kickoff locks", now: kickoff, want: true},
		{name: "after kickoff stays locked", now: kickoff.Add(time.Second), want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isLocked(tc.now, kickoff); got != tc.want {
				t.Fatalf("isLocked() = %v, want %v", got, tc.want)
			}
		})
	}
}
