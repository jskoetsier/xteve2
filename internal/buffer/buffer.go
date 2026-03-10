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
