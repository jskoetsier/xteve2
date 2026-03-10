// internal/hdhr/hdhr_test.go
package hdhr_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"xteve/internal/hdhr"
)

func TestDiscoverEndpoint(t *testing.T) {
	h := hdhr.New(hdhr.Config{
		DeviceID:   "12345678",
		TunerCount: 2,
		BaseURL:    "http://localhost:34400",
	})

	req := httptest.NewRequest("GET", "/discover.json", nil)
	w := httptest.NewRecorder()
	h.ServeDiscover(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var disc map[string]any
	if err := json.NewDecoder(w.Body).Decode(&disc); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if disc["DeviceID"] != "12345678" {
		t.Errorf("DeviceID = %v", disc["DeviceID"])
	}
	if disc["TunerCount"] != float64(2) {
		t.Errorf("TunerCount = %v", disc["TunerCount"])
	}
}

func TestLineupEndpoint(t *testing.T) {
	h := hdhr.New(hdhr.Config{
		DeviceID:   "12345678",
		TunerCount: 1,
		BaseURL:    "http://localhost:34400",
	})

	channels := []hdhr.LineupChannel{
		{GuideNumber: "1", GuideName: "CNN", URL: "http://localhost:34400/stream/abc"},
	}
	h.SetLineup(channels)

	req := httptest.NewRequest("GET", "/lineup.json", nil)
	w := httptest.NewRecorder()
	h.ServeLineup(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var lineup []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&lineup); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(lineup) != 1 {
		t.Fatalf("got %d entries, want 1", len(lineup))
	}
}
