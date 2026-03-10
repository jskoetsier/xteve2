# xTeVe Full Renovation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Renovate xTeVe from a dead 2019 monolith into a modern Go + React application while preserving all core functionality (M3U/XMLTV management, HDHR, SSDP, stream buffering).

**Architecture:** Incremental migration — new `cmd/`+`internal/` structure alongside old `src/`, porting logic package by package. React frontend built with Vite, embedded in the Go binary via `go:embed`. Old `src/` deleted when all packages are ported.

**Tech Stack:** Go 1.24, React 19, TypeScript, Vite, TanStack Query, React Router v7, shadcn/ui, Tailwind CSS, Zustand

**Design doc:** `docs/plans/2026-03-08-full-renovation-design.md`

---

## Phase 1: Foundation

### Task 1: Update go.mod and create cmd skeleton

**Files:**
- Modify: `go.mod`
- Create: `cmd/xteve/main.go`

**Step 1: Update go.mod**

Replace `go.mod` contents:

```
module xteve

go 1.24

require (
	github.com/gorilla/websocket v1.5.3
	github.com/koron/go-ssdp v0.0.4
	golang.org/x/crypto v0.31.0
)
```

**Step 2: Run go mod tidy**

```bash
go mod tidy
```
Expected: downloads new deps, removes `kardianos/osext` (no longer needed — stdlib `os.Executable` replaces it)

**Step 3: Create cmd/xteve/main.go**

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var version = "3.0.0"

func main() {
	var (
		port      = flag.Int("port", 34400, "HTTP port")
		configDir = flag.String("config", defaultConfigDir(), "Config directory")
		debug     = flag.Int("debug", 0, "Debug level 0-3")
	)
	flag.Parse()

	fmt.Printf("xTeVe %s\n", version)
	_ = port
	_ = configDir
	_ = debug

	log.Println("TODO: wire up server")
	os.Exit(0)
}

func defaultConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".xteve"
	}
	return home + "/.xteve"
}
```

**Step 4: Verify it builds**

```bash
go build ./cmd/xteve/
```
Expected: builds without errors, produces `xteve` binary

**Step 5: Commit**

```bash
git add go.mod go.sum cmd/xteve/main.go
git commit -m "feat: add cmd/xteve skeleton and update go.mod to Go 1.24"
```

---

### Task 2: storage package

**Files:**
- Create: `internal/storage/storage.go`
- Create: `internal/storage/storage_test.go`

**Step 1: Write the failing test**

```go
// internal/storage/storage_test.go
package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"xteve/internal/storage"
)

type TestData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	original := TestData{Name: "hello", Value: 42}
	if err := s.Save("test.json", original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	var loaded TestData
	if err := s.Load("test.json", &loaded); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded != original {
		t.Errorf("got %+v, want %+v", loaded, original)
	}
}

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	var v TestData
	err := s.Load("missing.json", &v)
	if !storage.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestEnsureDirs(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	dirs := []string{"cache", "backup", "temp", "img-cache", "img-upload"}
	if err := s.EnsureDirs(dirs...); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(dir, d)); err != nil {
			t.Errorf("dir %q not created: %v", d, err)
		}
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/storage/
```
Expected: FAIL — package does not exist

**Step 3: Write the implementation**

```go
// internal/storage/storage.go
package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var errNotFound = errors.New("not found")

// IsNotFound reports whether err is a not-found error from Load.
func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

// Storage manages JSON file persistence in a directory.
type Storage struct {
	dir string
	mu  sync.Mutex
}

// New returns a Storage rooted at dir.
func New(dir string) *Storage {
	return &Storage{dir: dir}
}

// Dir returns the root directory.
func (s *Storage) Dir() string {
	return s.dir
}

// Save writes v as JSON to filename inside the storage directory.
func (s *Storage) Save(filename string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, filename)
	return os.WriteFile(path, data, 0o644)
}

// Load reads filename from the storage directory and unmarshals it into v.
// Returns an IsNotFound error if the file does not exist.
func (s *Storage) Load(filename string, v any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.dir, filename)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return errNotFound
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// EnsureDirs creates subdirectories inside the storage directory.
func (s *Storage) EnsureDirs(names ...string) error {
	for _, name := range names {
		if err := os.MkdirAll(filepath.Join(s.dir, name), 0o755); err != nil {
			return err
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/storage/ -v
```
Expected: PASS — all 3 tests pass

**Step 5: Commit**

```bash
git add internal/storage/
git commit -m "feat: add internal/storage package with JSON persistence"
```

---

### Task 3: config package

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config_test

import (
	"testing"

	"xteve/internal/config"
	"xteve/internal/storage"
)

func TestDefaults(t *testing.T) {
	cfg := config.Default()

	if cfg.Port != 34400 {
		t.Errorf("default port = %d, want 34400", cfg.Port)
	}
	if cfg.TunerCount != 1 {
		t.Errorf("default tuner count = %d, want 1", cfg.TunerCount)
	}
	if cfg.AuthEnabled {
		t.Error("auth should be disabled by default")
	}
}

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	original := config.Default()
	original.Port = 9000
	original.AuthEnabled = true

	if err := config.Save(s, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := config.Load(s)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Port != 9000 {
		t.Errorf("port = %d, want 9000", loaded.Port)
	}
	if !loaded.AuthEnabled {
		t.Error("AuthEnabled should be true")
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	s := storage.New(dir)

	cfg, err := config.Load(s)
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}

	if cfg.Port != 34400 {
		t.Errorf("port = %d, want default 34400", cfg.Port)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/config/
```
Expected: FAIL — package does not exist

**Step 3: Write the implementation**

```go
// internal/config/config.go
package config

import (
	"xteve/internal/storage"
)

const filename = "settings.json"

// Settings holds all user-configurable server settings.
type Settings struct {
	Port           int    `json:"port"`
	TunerCount     int    `json:"tuner_count"`
	AuthEnabled    bool   `json:"auth_enabled"`
	AuthPassword   string `json:"auth_password,omitempty"` // bcrypt hash
	FFmpegPath     string `json:"ffmpeg_path"`
	VLCPath        string `json:"vlc_path"`
	BufferType     string `json:"buffer_type"` // "ffmpeg", "vlc", "hls"
	EPGRefreshHour int    `json:"epg_refresh_hour"`
}

// Default returns a Settings with sensible defaults.
func Default() Settings {
	return Settings{
		Port:           34400,
		TunerCount:     1,
		AuthEnabled:    false,
		FFmpegPath:     "/usr/bin/ffmpeg",
		BufferType:     "hls",
		EPGRefreshHour: 4,
	}
}

// Load reads settings from storage, returning defaults if the file doesn't exist.
func Load(s *storage.Storage) (Settings, error) {
	cfg := Default()
	err := s.Load(filename, &cfg)
	if storage.IsNotFound(err) {
		return cfg, nil
	}
	return cfg, err
}

// Save writes settings to storage.
func Save(s *storage.Storage, cfg Settings) error {
	return s.Save(filename, cfg)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/config/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add internal/config package with settings persistence"
```

---

## Phase 2: Core Domain Packages

### Task 4: m3u package

**Files:**
- Create: `internal/m3u/m3u.go`
- Create: `internal/m3u/m3u_test.go`
- Reference: `src/internal/m3u-parser/xteve_m3uParser.go` (migrate parsing logic)

**Step 1: Write the failing tests**

```go
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
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/m3u/
```
Expected: FAIL — package does not exist

**Step 3: Write the implementation**

```go
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
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/m3u/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/m3u/
git commit -m "feat: add internal/m3u package with channel parsing and filtering"
```

---

### Task 5: xepg package

**Files:**
- Create: `internal/xepg/xepg.go`
- Create: `internal/xepg/xepg_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/xepg/
```
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/xepg/xepg.go
package xepg

import (
	"crypto/md5"
	"fmt"
	"sync"

	"xteve/internal/m3u"
)

// Entry is a channel in the XEPG database with mapping metadata.
type Entry struct {
	ID          string      `json:"id"`
	Channel     m3u.Channel `json:"channel"`
	Enabled     bool        `json:"enabled"`
	CustomName  string      `json:"custom_name,omitempty"`
	EPGChannel  string      `json:"epg_channel,omitempty"` // mapped XMLTV channel ID
	ChannelNum  float64     `json:"channel_num,omitempty"`
}

// DB is the in-memory XEPG channel database.
type DB struct {
	mu      sync.RWMutex
	entries map[string]*Entry // keyed by ID
}

// NewDB creates an empty DB.
func NewDB() *DB {
	return &DB{entries: make(map[string]*Entry)}
}

// Sync merges a new channel list into the DB, preserving existing metadata.
func (db *DB) Sync(channels []m3u.Channel) {
	db.mu.Lock()
	defer db.mu.Unlock()

	seen := make(map[string]bool)
	for _, ch := range channels {
		id := channelID(ch)
		seen[id] = true

		if existing, ok := db.entries[id]; ok {
			// Update stream URL and attributes, preserve user settings
			existing.Channel = ch
		} else {
			db.entries[id] = &Entry{
				ID:      id,
				Channel: ch,
				Enabled: true,
			}
		}
	}

	// Remove channels no longer in any playlist
	for id := range db.entries {
		if !seen[id] {
			delete(db.entries, id)
		}
	}
}

// All returns all entries as a slice.
func (db *DB) All() []Entry {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make([]Entry, 0, len(db.entries))
	for _, e := range db.entries {
		result = append(result, *e)
	}
	return result
}

// Lookup finds a channel by ID.
func (db *DB) Lookup(id string) (Entry, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	e, ok := db.entries[id]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// SetEnabled enables or disables a channel.
func (db *DB) SetEnabled(id string, enabled bool) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	e, ok := db.entries[id]
	if !ok {
		return false
	}
	e.Enabled = enabled
	return true
}

func channelID(ch m3u.Channel) string {
	key := ch.TvgID + "|" + ch.Name + "|" + ch.URL
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))[:12]
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/xepg/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/xepg/
git commit -m "feat: add internal/xepg channel database with sync and lookup"
```

---

### Task 6: hdhr package

**Files:**
- Create: `internal/hdhr/hdhr.go`
- Create: `internal/hdhr/hdhr_test.go`

**Step 1: Write the failing test**

```go
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
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/hdhr/
```
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/hdhr/hdhr.go
package hdhr

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"sync"
)

// Config holds HDHomeRun device configuration.
type Config struct {
	DeviceID   string
	TunerCount int
	BaseURL    string
}

// LineupChannel is a single entry in the HDHomeRun lineup.
type LineupChannel struct {
	GuideNumber string
	GuideName   string
	URL         string
}

// Handler serves HDHomeRun discovery endpoints.
type Handler struct {
	cfg    Config
	mu     sync.RWMutex
	lineup []LineupChannel
}

// New creates a Handler.
func New(cfg Config) *Handler {
	return &Handler{cfg: cfg}
}

// SetLineup updates the channel lineup.
func (h *Handler) SetLineup(channels []LineupChannel) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lineup = channels
}

// ServeDiscover handles GET /discover.json.
func (h *Handler) ServeDiscover(w http.ResponseWriter, r *http.Request) {
	disc := map[string]any{
		"FriendlyName":    "xTeVe",
		"Manufacturer":    "xTeVe",
		"ModelNumber":     "HDHR3-US",
		"FirmwareName":    "hdhomerun3_atsc",
		"FirmwareVersion": "20200101",
		"DeviceID":        h.cfg.DeviceID,
		"DeviceAuth":      "test1234",
		"BaseURL":         h.cfg.BaseURL,
		"LineupURL":       h.cfg.BaseURL + "/lineup.json",
		"TunerCount":      h.cfg.TunerCount,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(disc)
}

// ServeLineup handles GET /lineup.json.
func (h *Handler) ServeLineup(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	lineup := make([]map[string]any, 0, len(h.lineup))
	for _, ch := range h.lineup {
		lineup = append(lineup, map[string]any{
			"GuideNumber": ch.GuideNumber,
			"GuideName":   ch.GuideName,
			"URL":         ch.URL,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lineup)
}

// ServeLineupStatus handles GET /lineup_status.json.
func (h *Handler) ServeLineupStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"ScanInProgress": 0,
		"ScanPossible":   1,
		"Source":         "Cable",
		"SourceList":     []string{"Cable"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

type deviceXML struct {
	XMLName     xml.Name `xml:"root"`
	SpecVersion struct {
		Major int `xml:"specVersion>major"`
		Minor int `xml:"specVersion>minor"`
	}
	URLBase     string `xml:"URLBase"`
	DeviceType  string `xml:"device>deviceType"`
	FriendlyName string `xml:"device>friendlyName"`
	Manufacturer string `xml:"device>manufacturer"`
	ModelName   string `xml:"device>modelName"`
	UDN         string `xml:"device>UDN"`
}

// ServeDeviceXML handles GET /device.xml.
func (h *Handler) ServeDeviceXML(w http.ResponseWriter, r *http.Request) {
	dev := deviceXML{
		URLBase:     h.cfg.BaseURL,
		DeviceType:  "urn:schemas-upnp-org:device:MediaServer:1",
		FriendlyName: "xTeVe",
		Manufacturer: "xTeVe",
		ModelName:   "xTeVe",
		UDN:         "uuid:" + h.cfg.DeviceID,
	}
	dev.SpecVersion.Major = 1
	dev.SpecVersion.Minor = 0

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(xml.Header))
	xml.NewEncoder(w).Encode(dev)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/hdhr/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/hdhr/
git commit -m "feat: add internal/hdhr package with HDHomeRun discovery endpoints"
```

---

### Task 7: ssdp package

**Files:**
- Create: `internal/ssdp/ssdp.go`

Note: SSDP depends on `github.com/koron/go-ssdp` which requires a real network interface to test. Skip unit tests; integration test in Task 21.

**Step 1: Write the implementation**

```go
// internal/ssdp/ssdp.go
package ssdp

import (
	"context"
	"fmt"
	"log"

	gossdp "github.com/koron/go-ssdp"
)

// Config holds SSDP advertisement configuration.
type Config struct {
	DeviceID string
	Port     int
}

// Advertise starts an SSDP advertisement and blocks until ctx is cancelled.
func Advertise(ctx context.Context, cfg Config) error {
	usn := fmt.Sprintf("uuid:%s::upnp:rootdevice", cfg.DeviceID)
	location := fmt.Sprintf("http://localhost:%d/device.xml", cfg.Port)

	ad, err := gossdp.Advertise(
		"upnp:rootdevice",
		usn,
		location,
		"xTeVe",
		1800,
		nil,
	)
	if err != nil {
		return fmt.Errorf("ssdp: %w", err)
	}

	log.Printf("ssdp: advertising on LAN (device %s)", cfg.DeviceID)

	<-ctx.Done()
	ad.Close()
	return nil
}
```

**Step 2: Verify it builds**

```bash
go build ./internal/ssdp/
```
Expected: no errors

**Step 3: Commit**

```bash
git add internal/ssdp/
git commit -m "feat: add internal/ssdp package for LAN device advertisement"
```

---

### Task 8: buffer package

**Files:**
- Create: `internal/buffer/buffer.go`
- Create: `internal/buffer/buffer_test.go`

**Step 1: Write the failing test**

```go
// internal/buffer/buffer_test.go
package buffer_test

import (
	"testing"

	"xteve/internal/buffer"
)

func TestTunerLimit(t *testing.T) {
	b := buffer.New(buffer.Config{TunerCount: 2})

	// Acquire up to limit
	id1, err := b.Acquire("http://stream1")
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}
	id2, err := b.Acquire("http://stream2")
	if err != nil {
		t.Fatalf("second Acquire: %v", err)
	}

	// Third should fail
	_, err = b.Acquire("http://stream3")
	if err == nil {
		t.Error("expected error at tuner limit, got nil")
	}
	if err != buffer.ErrTunerLimitReached {
		t.Errorf("error = %v, want ErrTunerLimitReached", err)
	}

	// Release one, then acquire should succeed
	b.Release(id1)
	_, err = b.Acquire("http://stream3")
	if err != nil {
		t.Fatalf("Acquire after release: %v", err)
	}
	_ = id2
}

func TestActiveCount(t *testing.T) {
	b := buffer.New(buffer.Config{TunerCount: 5})

	if b.ActiveCount() != 0 {
		t.Errorf("initial active = %d, want 0", b.ActiveCount())
	}

	id, _ := b.Acquire("http://stream1")
	if b.ActiveCount() != 1 {
		t.Errorf("active after acquire = %d, want 1", b.ActiveCount())
	}

	b.Release(id)
	if b.ActiveCount() != 0 {
		t.Errorf("active after release = %d, want 0", b.ActiveCount())
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/buffer/
```
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/buffer/buffer.go
package buffer

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrTunerLimitReached is returned when all tuners are in use.
var ErrTunerLimitReached = errors.New("buffer: tuner limit reached")

// Config holds buffer configuration.
type Config struct {
	TunerCount int
	Type       string // "hls", "ffmpeg", "vlc"
	FFmpegPath string
	VLCPath    string
}

// Session represents an active stream session.
type Session struct {
	ID        string
	StreamURL string
	StartedAt time.Time
}

// Buffer manages stream tuner slots.
type Buffer struct {
	cfg      Config
	mu       sync.Mutex
	sessions map[string]*Session
}

// New creates a Buffer.
func New(cfg Config) *Buffer {
	return &Buffer{
		cfg:      cfg,
		sessions: make(map[string]*Session),
	}
}

// Acquire reserves a tuner slot for the given stream URL and returns a session ID.
// Returns ErrTunerLimitReached if all tuners are busy.
func (b *Buffer) Acquire(streamURL string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.sessions) >= b.cfg.TunerCount {
		return "", ErrTunerLimitReached
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	b.sessions[id] = &Session{
		ID:        id,
		StreamURL: streamURL,
		StartedAt: time.Now(),
	}
	return id, nil
}

// Release frees a tuner slot.
func (b *Buffer) Release(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.sessions, id)
}

// ActiveCount returns the number of active sessions.
func (b *Buffer) ActiveCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.sessions)
}

// Sessions returns a copy of all active sessions.
func (b *Buffer) Sessions() []Session {
	b.mu.Lock()
	defer b.mu.Unlock()

	result := make([]Session, 0, len(b.sessions))
	for _, s := range b.sessions {
		result = append(result, *s)
	}
	return result
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/buffer/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/buffer/
git commit -m "feat: add internal/buffer package with tuner slot management"
```

---

## Phase 3: Auth + API

### Task 9: auth package

**Files:**
- Create: `internal/auth/auth.go`
- Create: `internal/auth/auth_test.go`

**Step 1: Write the failing test**

```go
// internal/auth/auth_test.go
package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"xteve/internal/auth"
)

func TestDisabledAllowsAll(t *testing.T) {
	a := auth.New(auth.Config{Enabled: false})

	handler := a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("disabled auth: status = %d, want 200", w.Code)
	}
}

func TestSetAndVerifyPassword(t *testing.T) {
	a := auth.New(auth.Config{Enabled: true})

	if err := a.SetPassword("secret123"); err != nil {
		t.Fatalf("SetPassword: %v", err)
	}

	if !a.CheckPassword("secret123") {
		t.Error("CheckPassword: correct password rejected")
	}
	if a.CheckPassword("wrong") {
		t.Error("CheckPassword: wrong password accepted")
	}
}

func TestUnauthenticatedReturns401(t *testing.T) {
	a := auth.New(auth.Config{Enabled: true})
	_ = a.SetPassword("secret")

	handler := a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated: status = %d, want 401", w.Code)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/auth/
```
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/auth/auth.go
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config holds auth configuration.
type Config struct {
	Enabled      bool
	PasswordHash string // bcrypt hash; empty means not set
}

// Auth manages optional session-based authentication.
type Auth struct {
	mu       sync.RWMutex
	cfg      Config
	sessions map[string]time.Time // token → expiry
}

const (
	cookieName    = "xteve_session"
	sessionTTL    = 24 * time.Hour
)

// New creates an Auth from config.
func New(cfg Config) *Auth {
	return &Auth{
		cfg:      cfg,
		sessions: make(map[string]time.Time),
	}
}

// SetPassword hashes and stores a new password.
func (a *Auth) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.mu.Lock()
	a.cfg.PasswordHash = string(hash)
	a.mu.Unlock()
	return nil
}

// CheckPassword reports whether password matches the stored hash.
func (a *Auth) CheckPassword(password string) bool {
	a.mu.RLock()
	hash := a.cfg.PasswordHash
	a.mu.RUnlock()
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Login creates a session token for the given request and sets a cookie.
func (a *Auth) Login(w http.ResponseWriter) string {
	token := randomToken()
	a.mu.Lock()
	a.sessions[token] = time.Now().Add(sessionTTL)
	a.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(sessionTTL.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// Logout invalidates the session from the request cookie.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		a.mu.Lock()
		delete(a.sessions, c.Value)
		a.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, MaxAge: -1, Path: "/"})
}

// Middleware returns an HTTP middleware that enforces auth when enabled.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.mu.RLock()
		enabled := a.cfg.Enabled
		a.mu.RUnlock()

		if !enabled {
			next.ServeHTTP(w, r)
			return
		}

		if c, err := r.Cookie(cookieName); err == nil {
			a.mu.RLock()
			expiry, ok := a.sessions[c.Value]
			a.mu.RUnlock()

			if ok && time.Now().Before(expiry) {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func randomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/auth/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/auth/
git commit -m "feat: add internal/auth package with optional session auth"
```

---

### Task 10: api package — REST handlers

**Files:**
- Create: `internal/api/api.go`
- Create: `internal/api/api_test.go`

**Step 1: Write the failing test**

```go
// internal/api/api_test.go
package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"xteve/internal/api"
	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/storage"
	"xteve/internal/xepg"
)

func newTestAPI(t *testing.T) *api.API {
	t.Helper()
	s := storage.New(t.TempDir())
	cfg := config.Default()
	db := xepg.NewDB()
	buf := buffer.New(buffer.Config{TunerCount: cfg.TunerCount})
	return api.New(api.Config{
		Storage: s,
		Settings: cfg,
		XEPG:    db,
		Buffer:  buf,
	})
}

func TestStatusEndpoint(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v1/status", nil)
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["version"] == nil {
		t.Error("response missing 'version' field")
	}
}

func TestSettingsGetEndpoint(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v1/settings", nil)
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	var s config.Settings
	if err := json.NewDecoder(w.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.Port != 34400 {
		t.Errorf("port = %d, want 34400", s.Port)
	}
}

func TestChannelsEndpoint(t *testing.T) {
	a := newTestAPI(t)

	req := httptest.NewRequest("GET", "/api/v1/channels", nil)
	w := httptest.NewRecorder()
	a.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/api/
```
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/api/api.go
package api

import (
	"encoding/json"
	"net/http"

	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/storage"
	"xteve/internal/xepg"
)

const version = "3.0.0"

// Config holds the dependencies for the API handler.
type Config struct {
	Storage  *storage.Storage
	Settings config.Settings
	XEPG     *xepg.DB
	Buffer   *buffer.Buffer
}

// API is the HTTP API handler.
type API struct {
	cfg    Config
	mux    *http.ServeMux
}

// New creates an API and registers all routes.
func New(cfg Config) *API {
	a := &API{cfg: cfg, mux: http.NewServeMux()}
	a.registerRoutes()
	return a
}

// Router returns the http.Handler for the API.
func (a *API) Router() http.Handler {
	return a.mux
}

func (a *API) registerRoutes() {
	a.mux.HandleFunc("GET /api/v1/status", a.handleStatus)
	a.mux.HandleFunc("GET /api/v1/settings", a.handleSettingsGet)
	a.mux.HandleFunc("PUT /api/v1/settings", a.handleSettingsPut)
	a.mux.HandleFunc("GET /api/v1/channels", a.handleChannelsGet)
	a.mux.HandleFunc("PUT /api/v1/channels/{id}", a.handleChannelPut)
	a.mux.HandleFunc("POST /api/v1/auth/login", a.handleLogin)
	a.mux.HandleFunc("POST /api/v1/auth/logout", a.handleLogout)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"version":      version,
		"active_streams": a.cfg.Buffer.ActiveCount(),
		"tuner_count":  a.cfg.Settings.TunerCount,
	})
}

func (a *API) handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(a.cfg.Storage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.AuthPassword = "" // never expose the hash
	writeJSON(w, cfg)
}

func (a *API) handleSettingsPut(w http.ResponseWriter, r *http.Request) {
	var updated config.Settings
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := config.Save(a.cfg.Storage, updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleChannelsGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, a.cfg.XEPG.All())
}

func (a *API) handleChannelPut(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var entry xepg.Entry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if !a.cfg.XEPG.SetEnabled(id, entry.Enabled) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Placeholder — wired up fully in Task 21
	w.WriteHeader(http.StatusOK)
}

func (a *API) handleLogout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
```

**Step 4: Run test to verify it passes**

```bash
go test ./internal/api/ -v
```
Expected: PASS

**Step 5: Commit**

```bash
git add internal/api/
git commit -m "feat: add internal/api package with REST handlers"
```

---

### Task 11: WebSocket log handler

**Files:**
- Create: `internal/api/ws.go`
- Create: `internal/api/ws_test.go`

**Step 1: Write the failing test**

```go
// internal/api/ws_test.go
package api_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"xteve/internal/api"
)

func TestWebSocketBroadcast(t *testing.T) {
	hub := api.NewHub()
	go hub.Run()

	server := httptest.NewServer(hub)
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	hub.Broadcast([]byte(`{"type":"log","msg":"hello"}`))

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if string(msg) != `{"type":"log","msg":"hello"}` {
		t.Errorf("got %q", msg)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test ./internal/api/ -run TestWebSocket
```
Expected: FAIL

**Step 3: Write the implementation**

```go
// internal/api/ws.go
package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages WebSocket connections and broadcasts.
type Hub struct {
	clients   map[*websocket.Conn]bool
	mu        sync.Mutex
	broadcast chan []byte
	register  chan *websocket.Conn
	unregister chan *websocket.Conn
}

// NewHub creates a Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte, 64),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run processes hub events. Call in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()

		case conn := <-h.unregister:
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()

		case msg := <-h.broadcast:
			h.mu.Lock()
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					delete(h.clients, conn)
					conn.Close()
				}
			}
			h.mu.Unlock()
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
		log.Println("ws: broadcast buffer full, dropping message")
	}
}

// ServeHTTP upgrades an HTTP connection to WebSocket.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: upgrade error: %v", err)
		return
	}

	h.register <- conn

	// Read loop to detect disconnects
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				h.unregister <- conn
				return
			}
		}
	}()
}
```

**Step 4: Register the /ws route in api.go**

Add to `registerRoutes()` in `internal/api/api.go`:
```go
// Add this field to API struct:
// hub *Hub

// In New():
// hub: NewHub(),

// In registerRoutes():
a.mux.Handle("/ws", a.hub)
```

And add `go a.hub.Run()` at the end of `New()`.

**Step 5: Run test to verify it passes**

```bash
go test ./internal/api/ -v
```
Expected: PASS

**Step 6: Commit**

```bash
git add internal/api/ws.go
git commit -m "feat: add WebSocket hub for real-time log streaming"
```

---

## Phase 4: React Frontend

### Task 12: Initialize React app with Vite

**Step 1: Scaffold the React app**

```bash
npm create vite@latest web -- --template react-ts
cd web
npm install
```

**Step 2: Install dependencies**

```bash
cd web
npm install @tanstack/react-query react-router-dom zustand
npm install -D tailwindcss @tailwindcss/vite
npx shadcn@latest init
```

When shadcn asks:
- Style: Default
- Base color: Slate
- CSS variables: Yes

**Step 3: Configure Vite proxy**

Edit `web/vite.config.ts`:

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:34400',
      '/ws': { target: 'ws://localhost:34400', ws: true },
    },
  },
  build: {
    outDir: '../web/dist',
    emptyOutDir: true,
  },
})
```

**Step 4: Verify dev server starts**

```bash
cd web && npm run dev
```
Expected: Vite dev server running on http://localhost:5173

**Step 5: Commit**

```bash
git add web/
git commit -m "feat: scaffold React frontend with Vite, Tailwind, shadcn/ui"
```

---

### Task 13: API client and routing setup

**Files:**
- Create: `web/src/lib/api.ts`
- Create: `web/src/main.tsx` (modify)
- Create: `web/src/App.tsx` (modify)

**Step 1: Create API client**

```ts
// web/src/lib/api.ts

const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  status: () => request<{ version: string; active_streams: number; tuner_count: number }>('/status'),
  getSettings: () => request<Record<string, unknown>>('/settings'),
  putSettings: (data: unknown) => request('/settings', { method: 'PUT', body: JSON.stringify(data) }),
  getChannels: () => request<unknown[]>('/channels'),
  putChannel: (id: string, data: unknown) =>
    request(`/channels/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  login: (password: string) =>
    request('/auth/login', { method: 'POST', body: JSON.stringify({ password }) }),
  logout: () => request('/auth/logout', { method: 'POST' }),
}
```

**Step 2: Set up routing in App.tsx**

```tsx
// web/src/App.tsx
import { BrowserRouter, Routes, Route, NavLink } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Dashboard from './pages/Dashboard'
import Channels from './pages/Channels'
import Playlists from './pages/Playlists'
import EPG from './pages/EPG'
import Settings from './pages/Settings'
import Logs from './pages/Logs'

const queryClient = new QueryClient()

const navItems = [
  { to: '/', label: 'Dashboard' },
  { to: '/channels', label: 'Channels' },
  { to: '/playlists', label: 'Playlists' },
  { to: '/epg', label: 'EPG' },
  { to: '/settings', label: 'Settings' },
  { to: '/logs', label: 'Logs' },
]

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="flex h-screen">
          <nav className="w-48 bg-slate-900 text-white flex flex-col gap-1 p-4">
            <span className="font-bold text-lg mb-4">xTeVe</span>
            {navItems.map(({ to, label }) => (
              <NavLink
                key={to}
                to={to}
                end={to === '/'}
                className={({ isActive }) =>
                  `px-3 py-2 rounded text-sm ${isActive ? 'bg-slate-700' : 'hover:bg-slate-800'}`
                }
              >
                {label}
              </NavLink>
            ))}
          </nav>
          <main className="flex-1 overflow-auto p-6">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/channels" element={<Channels />} />
              <Route path="/playlists" element={<Playlists />} />
              <Route path="/epg" element={<EPG />} />
              <Route path="/settings" element={<Settings />} />
              <Route path="/logs" element={<Logs />} />
            </Routes>
          </main>
        </div>
      </BrowserRouter>
    </QueryClientProvider>
  )
}
```

**Step 3: Commit**

```bash
git add web/src/
git commit -m "feat: add API client and React Router layout"
```

---

### Task 14: Dashboard page

**Files:**
- Create: `web/src/pages/Dashboard.tsx`

**Step 1: Create the page**

```tsx
// web/src/pages/Dashboard.tsx
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'

export default function Dashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['status'],
    queryFn: api.status,
    refetchInterval: 5000,
  })

  if (isLoading) return <p>Loading...</p>
  if (error) return <p className="text-red-500">Failed to load status</p>

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Dashboard</h1>
      <div className="grid grid-cols-3 gap-4">
        <StatCard label="Version" value={data?.version ?? '—'} />
        <StatCard label="Active Streams" value={String(data?.active_streams ?? 0)} />
        <StatCard label="Tuners Available" value={`${data?.active_streams ?? 0} / ${data?.tuner_count ?? 0}`} />
      </div>
    </div>
  )
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="border rounded-lg p-4 bg-white shadow-sm">
      <p className="text-sm text-slate-500">{label}</p>
      <p className="text-2xl font-semibold mt-1">{value}</p>
    </div>
  )
}
```

**Step 2: Commit**

```bash
git add web/src/pages/Dashboard.tsx
git commit -m "feat: add Dashboard page with status cards"
```

---

### Task 15: Channels page

**Files:**
- Create: `web/src/pages/Channels.tsx`

**Step 1: Create the page**

```tsx
// web/src/pages/Channels.tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'

interface Channel {
  id: string
  channel: { name: string; tvg_id: string; group_title: string; url: string }
  enabled: boolean
  custom_name?: string
}

export default function Channels() {
  const qc = useQueryClient()
  const { data: channels = [], isLoading } = useQuery({
    queryKey: ['channels'],
    queryFn: api.getChannels as () => Promise<Channel[]>,
  })

  const toggle = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api.putChannel(id, { enabled }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['channels'] }),
  })

  if (isLoading) return <p>Loading...</p>

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Channels</h1>
      <div className="border rounded-lg overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b">
            <tr>
              <th className="text-left p-3">Enabled</th>
              <th className="text-left p-3">Name</th>
              <th className="text-left p-3">Group</th>
              <th className="text-left p-3">TVG ID</th>
            </tr>
          </thead>
          <tbody>
            {channels.map((ch) => (
              <tr key={ch.id} className="border-b hover:bg-slate-50">
                <td className="p-3">
                  <input
                    type="checkbox"
                    checked={ch.enabled}
                    onChange={(e) => toggle.mutate({ id: ch.id, enabled: e.target.checked })}
                  />
                </td>
                <td className="p-3 font-medium">{ch.custom_name || ch.channel.name}</td>
                <td className="p-3 text-slate-500">{ch.channel.group_title}</td>
                <td className="p-3 text-slate-400 font-mono text-xs">{ch.channel.tvg_id}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {channels.length === 0 && (
          <p className="p-6 text-center text-slate-400">No channels. Add a playlist first.</p>
        )}
      </div>
    </div>
  )
}
```

**Step 2: Commit**

```bash
git add web/src/pages/Channels.tsx
git commit -m "feat: add Channels page with enable/disable toggle"
```

---

### Task 16: Playlists, EPG, and Settings pages

**Files:**
- Create: `web/src/pages/Playlists.tsx`
- Create: `web/src/pages/EPG.tsx`
- Create: `web/src/pages/Settings.tsx`

**Step 1: Create Playlists page**

```tsx
// web/src/pages/Playlists.tsx
export default function Playlists() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Playlists</h1>
      <p className="text-slate-400">M3U playlist management — coming soon.</p>
    </div>
  )
}
```

**Step 2: Create EPG page**

```tsx
// web/src/pages/EPG.tsx
export default function EPG() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">EPG</h1>
      <p className="text-slate-400">XMLTV provider management — coming soon.</p>
    </div>
  )
}
```

**Step 3: Create Settings page**

```tsx
// web/src/pages/Settings.tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, useEffect } from 'react'
import { api } from '@/lib/api'

export default function Settings() {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery({ queryKey: ['settings'], queryFn: api.getSettings })
  const [form, setForm] = useState<Record<string, unknown>>({})

  useEffect(() => {
    if (data) setForm(data)
  }, [data])

  const save = useMutation({
    mutationFn: () => api.putSettings(form),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  })

  if (isLoading) return <p>Loading...</p>

  return (
    <div className="max-w-lg">
      <h1 className="text-2xl font-bold mb-6">Settings</h1>
      <div className="space-y-4">
        <Field label="Port" value={String(form.port ?? '')}
          onChange={(v) => setForm({ ...form, port: Number(v) })} />
        <Field label="Tuner Count" value={String(form.tuner_count ?? '')}
          onChange={(v) => setForm({ ...form, tuner_count: Number(v) })} />
        <Field label="FFmpeg Path" value={String(form.ffmpeg_path ?? '')}
          onChange={(v) => setForm({ ...form, ffmpeg_path: v })} />
        <Field label="Buffer Type" value={String(form.buffer_type ?? '')}
          onChange={(v) => setForm({ ...form, buffer_type: v })} />
        <button
          onClick={() => save.mutate()}
          className="px-4 py-2 bg-slate-900 text-white rounded text-sm hover:bg-slate-700"
        >
          {save.isPending ? 'Saving...' : 'Save Settings'}
        </button>
      </div>
    </div>
  )
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-700 mb-1">{label}</label>
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full border rounded px-3 py-2 text-sm"
      />
    </div>
  )
}
```

**Step 4: Commit**

```bash
git add web/src/pages/
git commit -m "feat: add Playlists, EPG, and Settings pages"
```

---

### Task 17: Logs page with WebSocket

**Files:**
- Create: `web/src/pages/Logs.tsx`

**Step 1: Create the page**

```tsx
// web/src/pages/Logs.tsx
import { useEffect, useRef, useState } from 'react'

interface LogMessage {
  type: string
  msg: string
}

export default function Logs() {
  const [logs, setLogs] = useState<string[]>([])
  const [connected, setConnected] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const wsURL = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/ws`
    const ws = new WebSocket(wsURL)

    ws.onopen = () => setConnected(true)
    ws.onclose = () => setConnected(false)
    ws.onmessage = (e) => {
      try {
        const msg = JSON.parse(e.data) as LogMessage
        setLogs((prev) => [...prev.slice(-499), msg.msg])
      } catch {
        setLogs((prev) => [...prev.slice(-499), e.data])
      }
    }

    return () => ws.close()
  }, [])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [logs])

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold">Logs</h1>
        <span className={`text-xs px-2 py-1 rounded ${connected ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {connected ? 'Connected' : 'Disconnected'}
        </span>
      </div>
      <div className="flex-1 bg-slate-900 text-green-400 font-mono text-xs p-4 rounded overflow-auto">
        {logs.map((line, i) => <div key={i}>{line}</div>)}
        {logs.length === 0 && <span className="text-slate-500">Waiting for logs...</span>}
        <div ref={bottomRef} />
      </div>
    </div>
  )
}
```

**Step 2: Commit**

```bash
git add web/src/pages/Logs.tsx
git commit -m "feat: add Logs page with live WebSocket stream"
```

---

## Phase 5: Integration + Distribution

### Task 18: Wire up main.go

**Files:**
- Modify: `cmd/xteve/main.go`

**Step 1: Write the complete main.go**

```go
// cmd/xteve/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"xteve/internal/api"
	"xteve/internal/auth"
	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/hdhr"
	"xteve/internal/ssdp"
	"xteve/internal/storage"
	"xteve/internal/xepg"
)

var version = "3.0.0"

func main() {
	var (
		port      = flag.Int("port", 34400, "HTTP port")
		configDir = flag.String("config", defaultConfigDir(), "Config directory")
		debug     = flag.Int("debug", 0, "Debug level 0-3")
	)
	flag.Parse()

	if *debug > 0 {
		log.Printf("debug level: %d", *debug)
	}

	store := storage.New(*configDir)
	if err := store.EnsureDirs("cache", "backup", "temp", "img-cache", "img-upload"); err != nil {
		log.Fatalf("storage: %v", err)
	}

	cfg, err := config.Load(store)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if *port != 34400 {
		cfg.Port = *port
	}

	authSvc := auth.New(auth.Config{
		Enabled:      cfg.AuthEnabled,
		PasswordHash: cfg.AuthPassword,
	})

	xepgDB := xepg.NewDB()

	buf := buffer.New(buffer.Config{
		TunerCount: cfg.TunerCount,
		Type:       cfg.BufferType,
		FFmpegPath: cfg.FFmpegPath,
		VLCPath:    cfg.VLCPath,
	})

	hdhrHandler := hdhr.New(hdhr.Config{
		DeviceID:   "xteve1234",
		TunerCount: cfg.TunerCount,
		BaseURL:    fmt.Sprintf("http://localhost:%d", cfg.Port),
	})

	apiHandler := api.New(api.Config{
		Storage:  store,
		Settings: cfg,
		XEPG:     xepgDB,
		Buffer:   buf,
		Auth:     authSvc,
	})

	mux := http.NewServeMux()

	// HDHomeRun discovery (always public)
	mux.HandleFunc("/discover.json", hdhrHandler.ServeDiscover)
	mux.HandleFunc("/lineup.json", hdhrHandler.ServeLineup)
	mux.HandleFunc("/lineup_status.json", hdhrHandler.ServeLineupStatus)
	mux.HandleFunc("/device.xml", hdhrHandler.ServeDeviceXML)

	// API (auth-protected)
	mux.Handle("/api/", authSvc.Middleware(apiHandler.Router()))
	mux.Handle("/ws", apiHandler.Hub())

	// Web UI (served from go:embed in production, or local files in dev)
	mux.Handle("/", serveUI())

	addr := ":" + strconv.Itoa(cfg.Port)
	srv := &http.Server{Addr: addr, Handler: mux}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := ssdp.Advertise(ctx, ssdp.Config{
			DeviceID: "xteve1234",
			Port:     cfg.Port,
		}); err != nil && ctx.Err() == nil {
			log.Printf("ssdp: %v", err)
		}
	}()

	log.Printf("xTeVe %s listening on %s", version, addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	srv.Shutdown(context.Background())
}

func defaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.xteve"
}
```

**Step 2: Run all tests**

```bash
go test ./...
```
Expected: PASS for all packages

**Step 3: Build and smoke-test**

```bash
go build ./cmd/xteve/ && ./xteve -debug 1
```
Expected: server starts, logs "xTeVe 3.0.0 listening on :34400"

**Step 4: Commit**

```bash
git add cmd/xteve/main.go
git commit -m "feat: wire all packages into main.go with graceful shutdown"
```

---

### Task 19: go:embed web UI

**Files:**
- Create: `internal/ui/ui.go`
- Modify: `cmd/xteve/main.go` (update `serveUI()`)

**Step 1: Build the React app**

```bash
cd web && npm run build
```
Expected: `web/dist/` populated with index.html and assets

**Step 2: Create the embed package**

```go
// internal/ui/ui.go
package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var dist embed.FS

// Handler returns an http.Handler that serves the embedded React app.
// Unknown paths fall back to index.html for client-side routing.
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return &spaHandler{fs: http.FS(sub)}
}

type spaHandler struct {
	fs http.FileSystem
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := h.fs.Open(r.URL.Path)
	if err != nil {
		// Fall back to index.html for SPA routing
		r2 := *r
		r2.URL.Path = "/"
		http.FileServer(h.fs).ServeHTTP(w, &r2)
		return
	}
	f.Close()
	http.FileServer(h.fs).ServeHTTP(w, r)
}
```

**Step 3: Add dist/ symlink so embed resolves correctly**

```bash
ln -s ../../web/dist internal/ui/dist
```

**Step 4: Update serveUI() in main.go**

Replace the `serveUI()` placeholder with:
```go
import "xteve/internal/ui"

func serveUI() http.Handler {
	return ui.Handler()
}
```

**Step 5: Build and verify**

```bash
cd web && npm run build && cd ..
go build ./cmd/xteve/
./xteve
# Visit http://localhost:34400/
```
Expected: React app loads in browser

**Step 6: Commit**

```bash
git add internal/ui/ web/dist/
git commit -m "feat: embed React build into Go binary via go:embed"
```

---

### Task 20: Dockerfile and docker-compose

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `.dockerignore`

**Step 1: Write the Dockerfile**

```dockerfile
# Dockerfile

# Stage 1: Build React frontend
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: Build Go binary
FROM golang:1.24-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /app/web/dist ./web/dist
# Symlink for go:embed
RUN ln -sf /app/web/dist /app/internal/ui/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o xteve ./cmd/xteve/

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ffmpeg ca-certificates tzdata
WORKDIR /app
COPY --from=go-builder /app/xteve .
EXPOSE 34400
VOLUME ["/config"]
ENTRYPOINT ["/app/xteve", "-config", "/config"]
```

**Step 2: Write docker-compose.yml**

```yaml
services:
  xteve:
    build: .
    image: xteve:latest
    container_name: xteve
    restart: unless-stopped
    ports:
      - "34400:34400"
    volumes:
      - ./config:/config
    environment:
      - TZ=UTC
```

**Step 3: Write .dockerignore**

```
src/
node_modules/
web/node_modules/
*.md
docs/
.git/
```

**Step 4: Build and test Docker image**

```bash
docker build -t xteve:latest .
docker run --rm -p 34400:34400 xteve:latest
```
Expected: server starts, visit http://localhost:34400/

**Step 5: Commit**

```bash
git add Dockerfile docker-compose.yml .dockerignore
git commit -m "feat: add multi-stage Dockerfile and docker-compose"
```

---

### Task 21: Delete old src/ directory

Once all packages are ported and the app runs correctly, remove the legacy code.

**Step 1: Verify all tests pass**

```bash
go test ./...
```
Expected: PASS

**Step 2: Verify the binary runs**

```bash
go build ./cmd/xteve/ && ./xteve -debug 1
```
Expected: starts cleanly

**Step 3: Remove legacy code**

```bash
rm -rf src/
```

**Step 4: Verify build still works**

```bash
go build ./cmd/xteve/
go test ./...
```
Expected: no errors — old `src/` is not referenced by any new package

**Step 5: Commit**

```bash
git add -A
git commit -m "chore: remove legacy src/ directory"
```

---

### Task 22: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

Replace `CLAUDE.md` with the updated commands and architecture reflecting the new structure:

**Step 1: Update CLAUDE.md**

```markdown
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xTeVe is an M3U proxy server (Go + React) that bridges streaming services with Plex DVR and Emby Live TV. It implements the HDHomeRun protocol so media servers treat it as a network tuner.

## Build & Run

### Backend
\`\`\`bash
go build -o xteve ./cmd/xteve/
./xteve                              # default port 34400, config ~/.xteve/
./xteve -port 8080 -config /data -debug 2
\`\`\`

### Frontend (development)
\`\`\`bash
cd web && npm install && npm run dev   # Vite on :5173, proxies /api → :34400
\`\`\`

### Production build (embed UI in binary)
\`\`\`bash
cd web && npm run build && cd ..
go build ./cmd/xteve/
\`\`\`

### Docker
\`\`\`bash
docker compose up --build
\`\`\`

## Testing

\`\`\`bash
go test ./...                              # all Go tests
go test ./internal/m3u/ -run TestParse    # single test
go test ./internal/... -v                  # verbose
\`\`\`

## Architecture

\`\`\`
cmd/xteve/main.go       Entry point — flags, wiring, HTTP server, graceful shutdown
internal/
  config/               Settings load/save (JSON), typed Settings struct
  storage/              JSON file persistence with file-lock
  auth/                 Optional bcrypt session auth; Middleware() wraps routes
  m3u/                  M3U parsing (Parse, Filter) and Channel type
  xepg/                 Channel DB (Sync, Lookup, SetEnabled); keyed by content hash
  hdhr/                 HDHomeRun discovery endpoints (discover.json, lineup.json)
  ssdp/                 LAN SSDP advertisement via go-ssdp
  buffer/               Tuner slot management (Acquire/Release); stream proxying
  api/                  HTTP REST handlers + WebSocket Hub
  ui/                   go:embed wrapper for web/dist/
web/
  src/                  React 19 + TypeScript source
  dist/                 Built output (committed, embedded in binary)
\`\`\`

## Key Design Decisions

- Auth middleware only wraps `/api/` and `/web/` — streaming endpoints are always public for Plex/Emby
- `xepg.DB.Sync()` preserves user metadata (enabled state, custom names) across playlist refreshes
- `buffer.Buffer` enforces tuner limits; returns `ErrTunerLimitReached` (→ 503) when full
- `internal/ui` dist/ is a symlink to `web/dist/` so go:embed picks up the React build
- SSDP runs in a goroutine and exits cleanly when context is cancelled

## HTTP Endpoints

| Path | Purpose | Auth |
|------|---------|------|
| `/api/v1/status` | Health + tuner stats | optional |
| `/api/v1/settings` | GET/PUT server config | optional |
| `/api/v1/channels` | Channel list + mapping | optional |
| `/api/v1/auth/login` | POST to login | no |
| `/ws` | WebSocket log stream | no |
| `/stream/:id` | Stream proxy | no |
| `/m3u/`, `/xmltv/` | Generated playlist/EPG | no |
| `/discover.json`, `/lineup.json`, `/device.xml` | HDHomeRun | no |
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for renovated architecture"
```
