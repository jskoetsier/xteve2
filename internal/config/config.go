// internal/config/config.go
package config

import (
	"os"
	"strconv"

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
	M3UURL         string `json:"m3u_url,omitempty"`
	XMLTVURL       string `json:"xmltv_url,omitempty"`
	M3URefreshMins int    `json:"m3u_refresh_mins"`
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
		M3URefreshMins: 15,
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

// ApplyEnvOverrides fills runtime-configurable values from environment variables.
// Existing persisted settings win unless the field is empty or zero-valued.
func ApplyEnvOverrides(cfg Settings) Settings {
	if cfg.M3UURL == "" {
		cfg.M3UURL = os.Getenv("XTEVE_M3U_URL")
	}
	if cfg.XMLTVURL == "" {
		cfg.XMLTVURL = os.Getenv("XTEVE_XMLTV_URL")
	}
	if cfg.M3URefreshMins == 0 {
		if value := os.Getenv("XTEVE_M3U_REFRESH_MINS"); value != "" {
			if mins, err := strconv.Atoi(value); err == nil && mins > 0 {
				cfg.M3URefreshMins = mins
			}
		}
	}
	return cfg
}
