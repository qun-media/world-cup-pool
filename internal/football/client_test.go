package football

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFixturesForSeasonParsesAPIResponse(t *testing.T) {
	oldBaseURL := baseURL
	defer func() { baseURL = oldBaseURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/fixtures" {
			t.Fatalf("path = %s, want /fixtures", r.URL.Path)
		}
		if got := r.URL.Query().Get("league"); got != "1" {
			t.Fatalf("league query = %q, want 1", got)
		}
		if got := r.URL.Query().Get("season"); got != "2022" {
			t.Fatalf("season query = %q, want 2022", got)
		}
		if got := r.Header.Get("x-apisports-key"); got != "secret-key" {
			t.Fatalf("x-apisports-key = %q, want secret-key", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"errors": {},
			"results": 1,
			"response": [{
				"fixture": { "id": 42, "date": "2022-12-18T15:00:00Z", "status": { "short": "PEN" } },
				"league": { "round": "Final" },
				"teams": { "home": { "name": "Argentina" }, "away": { "name": "France" } },
				"goals": { "home": 3, "away": 3 },
				"score": {
					"fulltime": { "home": 2, "away": 2 },
					"extratime": { "home": 1, "away": 1 },
					"penalty": { "home": 4, "away": 2 }
				}
			}]
		}`))
	}))
	defer server.Close()
	baseURL = server.URL

	fixtures, err := New("secret-key").FixturesForSeason(context.Background(), 2022)
	if err != nil {
		t.Fatalf("FixturesForSeason() error = %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("len(fixtures) = %d, want 1", len(fixtures))
	}
	fixture := fixtures[0]
	if fixture.ID != 42 || fixture.Round != "Final" || fixture.Status != "PEN" {
		t.Fatalf("unexpected fixture metadata: %+v", fixture)
	}
	if fixture.HomeName != "Argentina" || fixture.AwayName != "France" {
		t.Fatalf("teams = %s/%s", fixture.HomeName, fixture.AwayName)
	}
	if !fixture.Finished() || fixture.Live() {
		t.Fatalf("status helpers wrong for %q", fixture.Status)
	}
	if *fixture.FTHome != 2 || *fixture.FTAway != 2 || *fixture.ETHome != 1 || *fixture.ETAway != 1 || *fixture.PenHome != 4 || *fixture.PenAway != 2 {
		t.Fatalf("scores parsed wrong: %+v", fixture)
	}
}

func TestFixturesForSeasonReportsAPIErrors(t *testing.T) {
	oldBaseURL := baseURL
	defer func() { baseURL = oldBaseURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":{"plan":"season not available"},"results":0,"response":[]}`))
	}))
	defer server.Close()
	baseURL = server.URL

	if _, err := New("secret-key").FixturesForSeason(context.Background(), 2026); err == nil {
		t.Fatal("FixturesForSeason() error = nil, want API error")
	}
}
