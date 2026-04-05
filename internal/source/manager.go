package source

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/hdhr"
	"xteve/internal/m3u"
	"xteve/internal/xepg"
)

// Manager synchronizes upstream IPTV data into xTeVe runtime state.
type Manager struct {
	client  *http.Client
	buffer  *buffer.Buffer
	hdhr    *hdhr.Handler
	baseURL string

	mu       sync.RWMutex
	settings config.Settings
	xepgDB   *xepg.DB
	channels map[string]xepg.Entry
}

// NewManager creates a source manager for a single M3U/XMLTV upstream.
func NewManager(settings config.Settings, xepgDB *xepg.DB, hdhrHandler *hdhr.Handler, buf *buffer.Buffer, baseURL string) *Manager {
	return &Manager{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		buffer:   buf,
		hdhr:     hdhrHandler,
		baseURL:  strings.TrimRight(baseURL, "/"),
		settings: settings,
		xepgDB:   xepgDB,
		channels: make(map[string]xepg.Entry),
	}
}

// Start performs an initial refresh and keeps the playlist fresh in the background.
func (m *Manager) Start(ctx context.Context) {
	if err := m.RefreshPlaylist(ctx); err != nil {
		log.Printf("source: initial playlist refresh failed: %v", err)
	}

	interval := time.Duration(m.currentSettings().M3URefreshMins) * time.Minute
	if interval <= 0 {
		interval = 15 * time.Minute
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.RefreshPlaylist(ctx); err != nil {
					log.Printf("source: playlist refresh failed: %v", err)
				}
			}
		}
	}()
}

// UpdateSettings replaces the runtime settings used by the manager.
func (m *Manager) UpdateSettings(settings config.Settings) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.settings = settings
}

// SyncLineup rebuilds the published channel lineup from the current XEPG state.
func (m *Manager) SyncLineup() {
	m.syncChannels()
}

// RefreshPlaylist fetches, parses, and publishes the configured M3U source.
func (m *Manager) RefreshPlaylist(ctx context.Context) error {
	settings := m.currentSettings()
	if settings.M3UURL == "" {
		return errors.New("m3u source URL is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, settings.M3UURL, nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("playlist fetch returned %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parsed, err := m3u.Parse(body)
	if err != nil {
		return err
	}

	m.xepgDB.Sync(parsed)
	m.syncChannels()
	return nil
}

// RefreshEPG fetches and parses the XMLTV source.
func (m *Manager) RefreshEPG(ctx context.Context) error {
	settings := m.currentSettings()
	if settings.XMLTVURL == "" {
		return errors.New("xmltv source URL is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, settings.XMLTVURL, nil)
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("xmltv fetch returned %s", resp.Status)
	}

	if err := m.xepgDB.ImportXMLTV(resp.Body); err != nil {
		return fmt.Errorf("xmltv parse: %w", err)
	}

	return nil
}

// ServeM3U returns an M3U playlist backed by xTeVe stream URLs.
func (m *Manager) ServeM3U(w http.ResponseWriter, r *http.Request) {
	entries := orderedEntries(m.entries())

	w.Header().Set("Content-Type", "audio/x-mpegurl")
	_, _ = io.WriteString(w, "#EXTM3U\n")
	for index, entry := range entries {
		if !entry.Enabled {
			continue
		}

		name := entry.Channel.Name
		if entry.CustomName != "" {
			name = entry.CustomName
		}

		line := fmt.Sprintf(
			"#EXTINF:-1 tvg-id=\"%s\" tvg-name=\"%s\" tvg-logo=\"%s\" group-title=\"%s\" channel-number=\"%s\",%s\n%s/stream/%s\n",
			entry.Channel.TvgID,
			name,
			entry.Channel.TvgLogo,
			entry.Channel.GroupTitle,
			guideNumberWithIndex(entry, index),
			name,
			m.baseURL,
			entry.ID,
		)
		_, _ = io.WriteString(w, line)
	}
}

// ServeXMLTV proxies the configured XMLTV upstream to clients like Jellyfin.
func (m *Manager) ServeXMLTV(w http.ResponseWriter, r *http.Request) {
	settings := m.currentSettings()
	if settings.XMLTVURL == "" {
		http.Error(w, "xmltv source URL is not configured", http.StatusNotFound)
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, settings.XMLTVURL, nil)
	if err != nil {
		http.Error(w, "invalid xmltv request", http.StatusInternalServerError)
		return
	}

	resp, err := m.client.Do(req)
	if err != nil {
		http.Error(w, "failed to fetch xmltv", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("xmltv upstream returned %s", resp.Status), http.StatusBadGateway)
		return
	}

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/xml")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, resp.Body)
}

// ServeStream proxies an upstream channel stream while enforcing tuner limits.
func (m *Manager) ServeStream(w http.ResponseWriter, r *http.Request) {
	entry, ok := m.lookupEntry(r.PathValue("id"))
	if !ok || !entry.Enabled {
		http.NotFound(w, r)
		return
	}

	sessionID, err := m.buffer.Acquire(entry.Channel.URL)
	if err != nil {
		if errors.Is(err, buffer.ErrTunerLimitReached) {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		http.Error(w, "failed to allocate tuner", http.StatusInternalServerError)
		return
	}
	defer m.buffer.Release(sessionID)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, entry.Channel.URL, nil)
	if err != nil {
		http.Error(w, "invalid upstream stream URL", http.StatusInternalServerError)
		return
	}

	resp, err := m.client.Do(req)
	if err != nil {
		http.Error(w, "failed to open upstream stream", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (m *Manager) syncChannels() {
	entries := orderedEntries(m.xepgDB.All())
	lineup := make([]hdhr.LineupChannel, 0, len(entries))
	cache := make(map[string]xepg.Entry, len(entries))

	for index, entry := range entries {
		cache[entry.ID] = entry
		if !entry.Enabled {
			continue
		}

		name := entry.Channel.Name
		if entry.CustomName != "" {
			name = entry.CustomName
		}

		lineup = append(lineup, hdhr.LineupChannel{
			GuideNumber: guideNumberWithIndex(entry, index),
			GuideName:   name,
			URL:         m.baseURL + "/stream/" + entry.ID,
		})
	}

	m.hdhr.SetLineup(lineup)

	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels = cache
}

func (m *Manager) currentSettings() config.Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

func (m *Manager) lookupEntry(id string) (xepg.Entry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.channels[id]
	return entry, ok
}

func (m *Manager) entries() []xepg.Entry {
	return m.xepgDB.All()
}

func guideNumberWithIndex(entry xepg.Entry, index int) string {
	if entry.ChannelNum > 0 {
		return strconv.FormatFloat(entry.ChannelNum, 'f', -1, 64)
	}
	return strconv.Itoa(index + 1)
}

var qualityPrefixPattern = regexp.MustCompile(`^(?:[0-9]{1,2}K|FHD|UHD|HD|SD)\s*\|\s*`)

var preferredChannelOrder = map[string]int{
	"NPO 1":               1,
	"NPO 2":               2,
	"NPO 3":               3,
	"RTL 4":               4,
	"RTL 5":               5,
	"SBS 6":               6,
	"RTL 7":               7,
	"RTL 8":               8,
	"NET 5":               9,
	"VERONICA DISNEY XD":  10,
	"SBS 9":               11,
	"RTL Z":               12,
	"ZIGGO TV":            13,
	"NPO 1 EXTRA":         14,
	"NPO 2 EXTRA":         15,
	"NPO NIEUWS":          16,
	"NPO POLITIEK":        17,
	"BBC FIRST":           18,
	"COMEDY CENTRAL":      19,
	"PARAMOUNT NETWORK":   20,
	"STAR CHANNEL":        21,
	"DISCOVERY CHANNEL":   22,
	"NATIONAL GEOGRAPHIC": 23,
	"HISTORY CHANNEL":     24,
	"ESPN 1":              30,
	"ESPN 2":              31,
	"ESPN 3":              32,
	"ESPN 4":              33,
	"ZIGGO SPORT":         34,
	"ZIGGO SPORT 1":       34,
	"ZIGGO SPORT 2":       35,
	"ZIGGO SPORT 3":       36,
	"ZIGGO SPORT 4":       37,
	"ZIGGO SPORT 5":       38,
	"ZIGGO SPORT 6":       39,
	"EUROSPORT 1":         40,
	"EUROSPORT 2":         41,
	"VIAPLAY FORMULE 1":   42,
	"FORMULE 1":           43,
	"F1 TV PRO":           44,
}

func orderedEntries(entries []xepg.Entry) []xepg.Entry {
	ordered := append([]xepg.Entry(nil), entries...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := buildChannelSortKey(ordered[i])
		right := buildChannelSortKey(ordered[j])

		if left.bucket != right.bucket {
			return left.bucket < right.bucket
		}
		if left.preferred != right.preferred {
			return left.preferred < right.preferred
		}
		if left.group != right.group {
			return left.group < right.group
		}
		if left.name != right.name {
			return left.name < right.name
		}
		return ordered[i].ID < ordered[j].ID
	})
	return ordered
}

type channelSortKey struct {
	bucket    int
	preferred int
	group     string
	name      string
}

func buildChannelSortKey(entry xepg.Entry) channelSortKey {
	name := canonicalChannelName(displayName(entry))
	group := canonicalGroup(entry.Channel.GroupTitle)
	preferred := 9999
	if rank, ok := preferredChannelOrder[name]; ok {
		preferred = rank
	}

	return channelSortKey{
		bucket:    channelBucket(group, name),
		preferred: preferred,
		group:     group,
		name:      name,
	}
}

func displayName(entry xepg.Entry) string {
	if entry.CustomName != "" {
		return entry.CustomName
	}
	return entry.Channel.Name
}

func canonicalChannelName(name string) string {
	name = strings.TrimSpace(strings.ToUpper(name))
	for {
		updated := qualityPrefixPattern.ReplaceAllString(name, "")
		if updated == name {
			break
		}
		name = updated
	}

	replacer := strings.NewReplacer(
		"NL | ", "",
		"UK | ", "",
		"UK| ", "",
		"US | ", "",
		"USA | ", "",
		"VERONICA / DISNEY XD", "VERONICA DISNEY XD",
		"DISCOVERY+ ", "",
		"  ", " ",
	)
	name = replacer.Replace(name)
	return strings.TrimSpace(name)
}

func canonicalGroup(group string) string {
	group = strings.TrimSpace(strings.ToUpper(group))
	replacer := strings.NewReplacer(
		"ᵁᴴᴰ", "UHD",
		"ⁿᵉʷ", "NEW",
		"DISCOVERY+ ", "DISCOVERY ",
		"  ", " ",
	)
	return strings.TrimSpace(replacer.Replace(group))
}

func channelBucket(group, name string) int {
	switch {
	case strings.Contains(group, "NEDERLAND"):
		return 0
	case strings.Contains(group, "ESPN") || strings.Contains(group, "ZIGGO SPORT") || strings.Contains(group, "FORMULE 1") || strings.Contains(group, "RACING") || strings.Contains(group, "SPORT"):
		if strings.HasPrefix(group, "NL |") || strings.Contains(group, "VIAPLAY") {
			return 1
		}
	case strings.HasPrefix(group, "NL |"):
		return 2
	case strings.HasPrefix(group, "UK |") || strings.HasPrefix(group, "UK|"):
		return 3
	case strings.HasPrefix(group, "US |") || strings.HasPrefix(group, "USA |"):
		return 4
	case strings.Contains(name, "NPO ") || strings.Contains(name, "RTL ") || strings.Contains(name, "SBS ") || strings.Contains(name, "ZIGGO ") || strings.Contains(name, "ESPN ") || strings.Contains(name, "EUROSPORT ") || strings.Contains(name, "BBC FIRST"):
		return 0
	default:
		return 5
	}
	return 5
}
