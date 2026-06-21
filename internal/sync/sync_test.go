package sync

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func testMatchRecord(stage string) *core.Record {
	collection := core.NewBaseCollection("matches")
	collection.Fields.Add(&core.TextField{Name: "stage", Max: 16})
	collection.Fields.Add(&core.TextField{Name: "status", Max: 16})
	collection.Fields.Add(&core.TextField{Name: "homeTeam", Max: 32})
	collection.Fields.Add(&core.TextField{Name: "awayTeam", Max: 32})
	collection.Fields.Add(&core.TextField{Name: "penWinner", Max: 32})
	collection.Fields.Add(&core.TextField{Name: "advancer", Max: 32})
	collection.Fields.Add(&core.NumberField{Name: "ftHome", OnlyInt: true})
	collection.Fields.Add(&core.NumberField{Name: "ftAway", OnlyInt: true})
	collection.Fields.Add(&core.NumberField{Name: "etHome", OnlyInt: true})
	collection.Fields.Add(&core.NumberField{Name: "etAway", OnlyInt: true})
	collection.Fields.Add(&core.NumberField{Name: "penHome", OnlyInt: true})
	collection.Fields.Add(&core.NumberField{Name: "penAway", OnlyInt: true})
	collection.Fields.Add(&core.DateField{Name: "finalizedAt"})

	record := core.NewRecord(collection)
	record.Set("stage", stage)
	record.Set("homeTeam", "home")
	record.Set("awayTeam", "away")
	return record
}

func TestWc26Unfinished(t *testing.T) {
	// time_elapsed is intentionally ignored — "has it kicked off?" is decided
	// from our own match kickoff, so the feed only tells us finished vs not.
	tests := []struct {
		name string
		g    wc26Game
		want bool
	}{
		{"finished", wc26Game{Finished: "TRUE", TimeElapsed: "finished"}, false},
		{"finished lowercase", wc26Game{Finished: "true"}, false},
		{"not started", wc26Game{Finished: "FALSE", TimeElapsed: "notstarted"}, true},
		{"first half", wc26Game{Finished: "FALSE", TimeElapsed: "45"}, true},
		{"empty elapsed", wc26Game{Finished: "FALSE", TimeElapsed: ""}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.g.unfinished(); got != tc.want {
				t.Fatalf("unfinished() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFetchWc26GamesFailover(t *testing.T) {
	const oneGame = `{"games":[{"id":"1","home_score":"1","away_score":"0","finished":"FALSE","home_team_name_en":"A","away_team_name_en":"B"}]}`

	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadGateway)
	}))
	defer down.Close()
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"games":[]}`)
	}))
	defer empty.Close()
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, oneGame)
	}))
	defer ok.Close()

	t.Run("falls back past a down mirror", func(t *testing.T) {
		games, src, err := fetchWc26GamesFrom(context.Background(), []string{down.URL, ok.URL})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if src != ok.URL {
			t.Fatalf("src = %q, want %q", src, ok.URL)
		}
		if len(games) != 1 || games[0].HomeScore != "1" {
			t.Fatalf("games = %+v, want one game with home_score 1", games)
		}
	})

	t.Run("treats an empty payload as a soft failure", func(t *testing.T) {
		_, src, err := fetchWc26GamesFrom(context.Background(), []string{empty.URL, ok.URL})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if src != ok.URL {
			t.Fatalf("src = %q, want %q (should skip the empty mirror)", src, ok.URL)
		}
	})

	t.Run("errors when every mirror fails", func(t *testing.T) {
		if _, _, err := fetchWc26GamesFrom(context.Background(), []string{down.URL, empty.URL}); err == nil {
			t.Fatal("expected an error when all mirrors fail")
		}
	})
}

func TestApplyResultStoresFinishedGroupResult(t *testing.T) {
	record := testMatchRecord("group")

	applyResult(record, "finished", pi(2), pi(1), nil, nil, nil, nil)

	if record.GetString("status") != "finished" {
		t.Fatalf("status = %q, want finished", record.GetString("status"))
	}
	if record.GetInt("ftHome") != 2 || record.GetInt("ftAway") != 1 {
		t.Fatalf("full-time score = %d-%d, want 2-1", record.GetInt("ftHome"), record.GetInt("ftAway"))
	}
	if record.GetInt("etHome") != 0 || record.GetInt("etAway") != 0 || record.GetInt("penHome") != 0 || record.GetInt("penAway") != 0 {
		t.Fatalf("extra-time/penalty defaults were not cleared")
	}
	if record.GetDateTime("finalizedAt").Time().IsZero() {
		t.Fatal("finalizedAt was not set")
	}
	if record.GetString("advancer") != "" {
		t.Fatalf("group advancer = %q, want empty", record.GetString("advancer"))
	}
}

func TestApplyResultDerivesKnockoutAdvancerFromPenalties(t *testing.T) {
	record := testMatchRecord("FINAL")

	applyResult(record, "finished", pi(1), pi(1), pi(2), pi(2), pi(4), pi(3))

	if record.GetString("advancer") != "home" {
		t.Fatalf("advancer = %q, want home", record.GetString("advancer"))
	}
	if record.GetString("penWinner") != "home" {
		t.Fatalf("penWinner = %q, want home", record.GetString("penWinner"))
	}
}

func TestApplyResultDoesNotDeriveAdvancerBeforeFinished(t *testing.T) {
	record := testMatchRecord("R32")

	applyResult(record, "live", pi(2), pi(0), nil, nil, nil, nil)

	if record.GetString("advancer") != "" {
		t.Fatalf("advancer = %q, want empty while live", record.GetString("advancer"))
	}
	if !record.GetDateTime("finalizedAt").Time().IsZero() {
		t.Fatal("finalizedAt was set for a live match")
	}
}
