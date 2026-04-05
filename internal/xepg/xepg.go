// internal/xepg/xepg.go
package xepg

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"xteve/internal/m3u"
)

// Entry is a channel in the XEPG database with mapping metadata.
type Entry struct {
	ID         string      `json:"id"`
	Channel    m3u.Channel `json:"channel"`
	Enabled    bool        `json:"enabled"`
	CustomName string      `json:"custom_name"`
	EPGChannel string      `json:"epg_channel"`
	ChannelNum float64     `json:"channel_num"`
}

// Program is a single programme entry from an XMLTV file.
type Program struct {
	Channel  string    `json:"channel"`
	Start    time.Time `json:"start"`
	Stop     time.Time `json:"stop"`
	Title    string    `json:"title"`
	Desc     string    `json:"desc,omitempty"`
	Category string    `json:"category,omitempty"`
	Icon     string    `json:"icon,omitempty"`
	Episode  string    `json:"episode,omitempty"`
}

// DB is the in-memory XEPG channel database.
type DB struct {
	mu       sync.RWMutex
	entries  map[string]*Entry
	programs map[string][]Program // keyed by channel ID
}

// NewDB creates an empty DB.
func NewDB() *DB {
	return &DB{
		entries:  make(map[string]*Entry),
		programs: make(map[string][]Program),
	}
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
			existing.Channel = ch
		} else {
			db.entries[id] = &Entry{
				ID:      id,
				Channel: ch,
				Enabled: true,
			}
		}
	}

	for id := range db.entries {
		if !seen[id] {
			delete(db.entries, id)
			delete(db.programs, id)
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

// SetCustomName updates the custom name for a channel.
func (db *DB) SetCustomName(id string, name string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	e, ok := db.entries[id]
	if !ok {
		return false
	}
	e.CustomName = name
	return true
}

// SetEPGChannel maps an EPG channel ID to a channel.
func (db *DB) SetEPGChannel(id string, epgChannel string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	e, ok := db.entries[id]
	if !ok {
		return false
	}
	e.EPGChannel = epgChannel
	return true
}

// SetChannelNum sets the channel number.
func (db *DB) SetChannelNum(id string, num float64) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	e, ok := db.entries[id]
	if !ok {
		return false
	}
	e.ChannelNum = num
	return true
}

// ProgramsFor returns all programs for a given channel ID.
func (db *DB) ProgramsFor(channelID string) []Program {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.programs[channelID]
}

// AllPrograms returns all programs grouped by channel ID.
func (db *DB) AllPrograms() map[string][]Program {
	db.mu.RLock()
	defer db.mu.RUnlock()
	result := make(map[string][]Program, len(db.programs))
	for k, v := range db.programs {
		result[k] = v
	}
	return result
}

// SetPrograms replaces the program list for a channel.
func (db *DB) SetPrograms(channelID string, programs []Program) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.programs[channelID] = programs
}

// ImportXMLTV parses an XMLTV io.Reader and stores the programmes.
// It maps XMLTV channel IDs to xTeVe channel IDs via the EPGChannel field,
// and falls back to matching by tvg-id or display-name.
func (db *DB) ImportXMLTV(r io.Reader) error {
	dec := xml.NewDecoder(r)
	var tv XMLTV
	if err := dec.Decode(&tv); err != nil {
		return fmt.Errorf("xmltv decode: %w", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.programs = make(map[string][]Program)

	// Build a lookup: XMLTV channel ID -> xTeVe channel ID
	xmltvToXTeVe := make(map[string]string)

	// Build EPG channel display name -> XMLTV channel ID for fuzzy matching
	epgDisplayName := make(map[string]string)
	for _, ch := range tv.Channel {
		name := ""
		if len(ch.Display) > 0 {
			name = ch.Display[0].Value
		}
		epgDisplayName[strings.ToLower(name)] = ch.ID
	}

	// Map xTeVe entries to XMLTV channels
	for _, entry := range db.entries {
		matchedXMLTVID := ""

		// 1. User-configured EPGChannel mapping
		if entry.EPGChannel != "" {
			if _, ok := epgDisplayName[strings.ToLower(entry.EPGChannel)]; ok {
				matchedXMLTVID = epgDisplayName[strings.ToLower(entry.EPGChannel)]
			}
		}

		// 2. Match by tvg-id
		if matchedXMLTVID == "" && entry.Channel.TvgID != "" {
			for xmltvID := range epgDisplayName {
				if strings.EqualFold(entry.Channel.TvgID, xmltvID) {
					matchedXMLTVID = epgDisplayName[xmltvID]
					break
				}
			}
		}

		// 3. Match by channel name
		if matchedXMLTVID == "" {
			for xmltvName, xmltvID := range epgDisplayName {
				if strings.Contains(strings.ToLower(entry.Channel.Name), xmltvName) ||
					strings.Contains(xmltvName, strings.ToLower(entry.Channel.Name)) {
					matchedXMLTVID = xmltvID
					break
				}
			}
		}

		if matchedXMLTVID != "" {
			xmltvToXTeVe[matchedXMLTVID] = entry.ID
		}
	}

	// Group programmes by xTeVe channel ID
	for _, p := range tv.Programme {
		xTeVeID := xmltvToXTeVe[p.Channel]
		if xTeVeID == "" {
			continue
		}
		db.programs[xTeVeID] = append(db.programs[xTeVeID], p.ToProgram())
	}

	return nil
}

// XMLTV is the top-level XMLTV document structure.
type XMLTV struct {
	XMLName   xml.Name    `xml:"tv"`
	Channel   []Channel   `xml:"channel"`
	Programme []Programme `xml:"programme"`
}

// Channel represents a <channel> element.
type Channel struct {
	ID      string    `xml:"id"`
	Display []Display `xml:"display-name"`
	Icon    []Icon    `xml:"icon"`
}

// Display is a <display-name> element.
type Display struct {
	Value string `xml:",chardata"`
}

// Icon is an <icon> element.
type Icon struct {
	Src string `xml:"src,attr"`
}

// Programme represents a <programme> element.
type Programme struct {
	Channel  string     `xml:"channel,attr"`
	Start    string     `xml:"start,attr"`
	Stop     string     `xml:"stop,attr"`
	Title    []Title    `xml:"title"`
	Desc     []Desc     `xml:"desc"`
	Category []Category `xml:"category"`
	Icon     []Icon     `xml:"icon"`
	Episode  []Episode  `xml:"episode-num"`
}

// Title is a <title> element.
type Title struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

// Desc is a <desc> element.
type Desc struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

// Category is a <category> element.
type Category struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

// Icon is an <icon> element.
type IconProg struct {
	Src string `xml:"src,attr"`
}

// Episode is an <episode-num> element.
type Episode struct {
	System string `xml:"system,attr"`
	Value  string `xml:",chardata"`
}

// ToProgram converts an XMLTV Programme to a Program.
func (p Programme) ToProgram() Program {
	title := ""
	if len(p.Title) > 0 {
		title = p.Title[0].Value
	}
	desc := ""
	if len(p.Desc) > 0 {
		desc = p.Desc[0].Value
	}
	category := ""
	if len(p.Category) > 0 {
		category = p.Category[0].Value
	}
	icon := ""
	if len(p.Icon) > 0 {
		icon = p.Icon[0].Src
	}
	episode := ""
	if len(p.Episode) > 0 {
		episode = p.Episode[0].Value
	}

	start, _ := parseXMLTVDate(p.Start)
	stop, _ := parseXMLTVDate(p.Stop)

	return Program{
		Channel:  p.Channel,
		Start:    start,
		Stop:     stop,
		Title:    title,
		Desc:     desc,
		Category: category,
		Icon:     icon,
		Episode:  episode,
	}
}

// parseXMLTVDate parses XMLTV date format (YYYYMMDDHHMMSS or YYYYMMDDHHMMSS +ZZZZ).
func parseXMLTVDate(s string) (time.Time, error) {
	if len(s) < 14 {
		return time.Time{}, fmt.Errorf("invalid xmltv date: %s", s)
	}
	layout := "20060102150405 -0700"
	if len(s) == 14 {
		layout = "20060102150405"
	}
	return time.Parse(layout, s)
}

func channelID(ch m3u.Channel) string {
	key := ch.TvgID + "|" + ch.Name + "|" + ch.URL
	return fmt.Sprintf("%x", md5.Sum([]byte(key)))[:12]
}
