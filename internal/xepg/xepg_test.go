// internal/xepg/xepg_test.go
package xepg_test

import (
	"testing"

	"xteve/internal/m3u"
	"xteve/internal/xepg"
)

func TestMap(t *testing.T) {
	channels := []m3u.Channel{
		{Name: "CNN", TvgID: "cnn", GroupTitle: "News", URL: "http://stream/cnn"},
		{Name: "BBC", TvgID: "bbc", GroupTitle: "News", URL: "http://stream/bbc"},
	}

	db := xepg.NewDB()
	db.Sync(channels)

	entries := db.All()
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
}

func TestLookup(t *testing.T) {
	channels := []m3u.Channel{
		{Name: "CNN", TvgID: "cnn", URL: "http://stream/cnn"},
	}

	db := xepg.NewDB()
	db.Sync(channels)

	entries := db.All()
	if len(entries) == 0 {
		t.Fatal("no entries")
	}

	id := entries[0].ID
	entry, ok := db.Lookup(id)
	if !ok {
		t.Fatalf("Lookup(%q) not found", id)
	}
	if entry.Channel.URL != "http://stream/cnn" {
		t.Errorf("URL = %q", entry.Channel.URL)
	}
}

func TestSyncPreservesEnabled(t *testing.T) {
	ch := m3u.Channel{Name: "CNN", TvgID: "cnn", URL: "http://stream/cnn"}
	db := xepg.NewDB()
	db.Sync([]m3u.Channel{ch})

	// Disable channel
	entries := db.All()
	db.SetEnabled(entries[0].ID, false)

	// Re-sync (simulating playlist refresh)
	db.Sync([]m3u.Channel{ch})

	entries = db.All()
	if entries[0].Enabled {
		t.Error("re-sync should preserve disabled state")
	}
}
