package source

import (
	"testing"

	"xteve/internal/m3u"
	"xteve/internal/xepg"
)

func TestOrderedEntriesPrefersDutchCoreChannels(t *testing.T) {
	entries := []xepg.Entry{
		{
			ID: "uk1",
			Channel: m3u.Channel{
				Name:       "UK | SKY ATLANTIC",
				GroupTitle: "UK | ENTERTAINMENT",
			},
			Enabled: true,
		},
		{
			ID: "rtl4",
			Channel: m3u.Channel{
				Name:       "4K | RTL 4",
				GroupTitle: "NL | 4K NEDERLAND",
			},
			Enabled: true,
		},
		{
			ID: "espn1",
			Channel: m3u.Channel{
				Name:       "4K | ESPN 1",
				GroupTitle: "NL | ESPN & ZIGGO SPORT",
			},
			Enabled: true,
		},
		{
			ID: "npo1",
			Channel: m3u.Channel{
				Name:       "4K | NPO 1",
				GroupTitle: "NL | 4K NEDERLAND",
			},
			Enabled: true,
		},
	}

	ordered := orderedEntries(entries)
	got := []string{
		ordered[0].Channel.Name,
		ordered[1].Channel.Name,
		ordered[2].Channel.Name,
		ordered[3].Channel.Name,
	}

	want := []string{
		"4K | NPO 1",
		"4K | RTL 4",
		"4K | ESPN 1",
		"UK | SKY ATLANTIC",
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestGuideNumberWithIndexUsesOrderedFallback(t *testing.T) {
	entry := xepg.Entry{}
	if got := guideNumberWithIndex(entry, 6); got != "7" {
		t.Fatalf("guideNumberWithIndex() = %q, want 7", got)
	}
}
