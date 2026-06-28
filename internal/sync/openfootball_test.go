package sync

import "testing"

func TestKoCorrection(t *testing.T) {
	names := map[string]string{
		"Germany":  "ger",
		"Paraguay": "par",
	}
	tests := []struct {
		name         string
		t1, t2       string
		finished     bool
		wantOK       bool
		wantH, wantA string
	}{
		{"both real, not played", "Germany", "Paraguay", false, true, "ger", "par"},
		{"placeholder side", "Germany", "3A/B/C/D/F", false, false, "", ""},
		{"both placeholders", "W74", "W77", false, false, "", ""},
		{"already played is skipped", "Germany", "Paraguay", true, false, "", ""},
		{"unknown name skipped", "Germany", "Atlantis", false, false, "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, a, ok := koCorrection(tc.t1, tc.t2, names, tc.finished)
			if ok != tc.wantOK || h != tc.wantH || a != tc.wantA {
				t.Fatalf("koCorrection(%q,%q,finished=%v) = (%q,%q,%v), want (%q,%q,%v)",
					tc.t1, tc.t2, tc.finished, h, a, ok, tc.wantH, tc.wantA, tc.wantOK)
			}
		})
	}
}
