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
	ID         string      `json:"id"`
	Channel    m3u.Channel `json:"channel"`
	Enabled    bool        `json:"enabled"`
	CustomName string      `json:"custom_name,omitempty"`
	EPGChannel string      `json:"epg_channel,omitempty"` // mapped XMLTV channel ID
	ChannelNum float64     `json:"channel_num,omitempty"`
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
