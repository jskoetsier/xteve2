package source

import (
	"strings"
	"testing"
)

func TestStripChannelIconsFromXMLTV(t *testing.T) {
	const in = `<?xml version="1.0"?>
<tv>
  <channel id="a">
    <display-name>NPO 1</display-name>
    <icon src="https://wrong.example/a.png"/>
  </channel>
  <channel id="b">
    <display-name>BBC</display-name>
    <icon src="https://x.example/b.png"></icon>
  </channel>
  <programme channel="a" start="20260101000000">
    <title>Show</title>
    <icon src="https://keep.programme.example/p.png"/>
  </programme>
</tv>`

	out := stripChannelIconsFromXMLTV([]byte(in))
	s := string(out)
	if strings.Contains(s, "wrong.example") || strings.Contains(s, "x.example/b.png") {
		t.Fatalf("channel icons should be stripped:\n%s", s)
	}
	if !strings.Contains(s, "keep.programme.example") {
		t.Fatalf("programme icon should remain:\n%s", s)
	}
	if !strings.Contains(s, "<channel id=\"a\">") {
		t.Fatalf("expected channel block to remain:\n%s", s)
	}
}
