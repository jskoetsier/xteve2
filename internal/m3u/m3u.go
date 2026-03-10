// internal/m3u/m3u.go
package m3u

import (
	"errors"
	"regexp"
	"strings"
)

// Channel represents a single entry parsed from an M3U playlist.
type Channel struct {
	Name       string
	TvgID      string
	TvgName    string
	TvgLogo    string
	GroupTitle string
	URL        string
	Attrs      map[string]string // any additional attributes
}

var (
	reAttr = regexp.MustCompile(`([\w-]+)="([^"]*)"`)
	reURL  = regexp.MustCompile(`^https?://|^rtsp://|^rtmp://|^udp://`)
)

// Parse parses an M3U playlist from raw bytes and returns all channels.
func Parse(data []byte) ([]Channel, error) {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(content, "\n")

	if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "#EXTM3U") {
		return nil, errors.New("m3u: invalid format: missing #EXTM3U header")
	}

	var channels []Channel
	var pending *Channel

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			ch := parseExtInf(line)
			pending = &ch
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if reURL.MatchString(line) {
			if pending != nil {
				pending.URL = line
				channels = append(channels, *pending)
				pending = nil
			}
		}
	}

	return channels, nil
}

func parseExtInf(line string) Channel {
	ch := Channel{Attrs: make(map[string]string)}

	// Extract all key="value" pairs
	for _, m := range reAttr.FindAllStringSubmatch(line, -1) {
		key, val := strings.ToLower(m[1]), m[2]
		ch.Attrs[key] = val
		switch key {
		case "tvg-id":
			ch.TvgID = val
		case "tvg-name":
			ch.TvgName = val
		case "tvg-logo":
			ch.TvgLogo = val
		case "group-title":
			ch.GroupTitle = val
		}
	}

	// Channel display name is after the last comma
	if idx := strings.LastIndex(line, ","); idx != -1 {
		ch.Name = strings.TrimSpace(line[idx+1:])
	}

	return ch
}

// Filter returns channels matching the predicate.
func Filter(channels []Channel, keep func(Channel) bool) []Channel {
	result := make([]Channel, 0, len(channels))
	for _, ch := range channels {
		if keep(ch) {
			result = append(result, ch)
		}
	}
	return result
}
