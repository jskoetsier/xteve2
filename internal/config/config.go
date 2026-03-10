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
