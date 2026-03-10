// internal/m3u/m3u_test.go
package m3u_test

import (
	"testing"

	"xteve/internal/m3u"
)

var sampleM3U = `#EXTM3U
#EXTINF:-1 tvg-id="cnn" tvg-name="CNN" tvg-logo="https://example.com/cnn.png" group-title="News",CNN
http://stream.example.com/cnn
#EXTINF:-1 tvg-id="bbc" tvg-name="BBC World" tvg-logo="" group-title="News",BBC World News
http://stream.example.com/bbc
`

func TestParse(t *testing.T) {
	channels, err := m3u.Parse([]byte(sampleM3U))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(channels) != 2 {
		t.Fatalf("got %d channels, want 2", len(channels))
	}

	cnn := channels[0]
	if cnn.Name != "CNN" {
		t.Errorf("Name = %q, want CNN", cnn.Name)
	}
	if cnn.TvgID != "cnn" {
		t.Errorf("TvgID = %q, want cnn", cnn.TvgID)
	}
	if cnn.URL != "http://stream.example.com/cnn" {
		t.Errorf("URL = %q", cnn.URL)
	}
	if cnn.GroupTitle != "News" {
		t.Errorf("GroupTitle = %q, want News", cnn.GroupTitle)
	}
}

func TestParseEmpty(t *testing.T) {
	channels, err := m3u.Parse([]byte("#EXTM3U\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(channels) != 0 {
		t.Errorf("got %d channels, want 0", len(channels))
	}
}

func TestParseInvalidHeader(t *testing.T) {
	_, err := m3u.Parse([]byte("not an m3u file"))
	if err == nil {
		t.Error("expected error for invalid M3U, got nil")
	}
}

func TestFilter(t *testing.T) {
	channels, _ := m3u.Parse([]byte(sampleM3U))

	news := m3u.Filter(channels, func(c m3u.Channel) bool {
		return c.GroupTitle == "News"
	})
	if len(news) != 2 {
		t.Errorf("got %d channels after filter, want 2", len(news))
	}

	bbc := m3u.Filter(channels, func(c m3u.Channel) bool {
		return c.TvgID == "bbc"
	})
	if len(bbc) != 1 {
		t.Errorf("got %d channels after ID filter, want 1", len(bbc))
	}
}
