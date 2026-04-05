// internal/api/api.go
package api

import (
	"context"
	"encoding/json"
	"net/http"

	"xteve/internal/buffer"
	"xteve/internal/config"
	"xteve/internal/source"
	"xteve/internal/storage"
	"xteve/internal/xepg"
)

const version = "2.0.0"

// Config holds the dependencies for the API handler.
type Config struct {
	Storage           *storage.Storage
	Settings          config.Settings
	XEPG              *xepg.DB
	Buffer            *buffer.Buffer
	SourceManager     *source.Manager
	OnSettingsChanged func(config.Settings)
	OnChannelsChanged func()
}

// API is the HTTP API handler.
type API struct {
	cfg Config
	mux *http.ServeMux
	hub *Hub
}

// New creates an API and registers all routes.
func New(cfg Config) *API {
	a := &API{cfg: cfg, mux: http.NewServeMux(), hub: NewHub()}
	go a.hub.Run()
	a.registerRoutes()
	return a
}

// Router returns the http.Handler for the API.
func (a *API) Router() http.Handler {
	return a.mux
}

// Hub returns the WebSocket hub.
func (a *API) Hub() *Hub {
	return a.hub
}

func (a *API) registerRoutes() {
	a.mux.HandleFunc("GET /api/v1/status", a.handleStatus)
	a.mux.HandleFunc("GET /api/v1/settings", a.handleSettingsGet)
	a.mux.HandleFunc("PUT /api/v1/settings", a.handleSettingsPut)
	a.mux.HandleFunc("GET /api/v1/channels", a.handleChannelsGet)
	a.mux.HandleFunc("PUT /api/v1/channels/{id}", a.handleChannelPut)
	a.mux.HandleFunc("POST /api/v1/auth/login", a.handleLogin)
	a.mux.HandleFunc("POST /api/v1/auth/logout", a.handleLogout)
	a.mux.HandleFunc("POST /api/v1/playlists/refresh", a.handlePlaylistRefresh)
	a.mux.HandleFunc("POST /api/v1/epg/refresh", a.handleEPGRefresh)
	a.mux.HandleFunc("GET /api/v1/epg/programs", a.handleEPGPrograms)
	a.mux.HandleFunc("PUT /api/v1/channels/{id}/mapping", a.handleChannelMapping)
	a.mux.Handle("/ws", a.hub)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"version":        version,
		"active_streams": a.cfg.Buffer.ActiveCount(),
		"tuner_count":    a.cfg.Settings.TunerCount,
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
	a.cfg.Settings = updated
	if a.cfg.OnSettingsChanged != nil {
		a.cfg.OnSettingsChanged(updated)
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
	if a.cfg.OnChannelsChanged != nil {
		a.cfg.OnChannelsChanged()
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

func (a *API) handlePlaylistRefresh(w http.ResponseWriter, r *http.Request) {
	if a.cfg.SourceManager == nil {
		http.Error(w, "source manager not configured", http.StatusServiceUnavailable)
		return
	}
	if err := a.cfg.SourceManager.RefreshPlaylist(context.Background()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (a *API) handleEPGRefresh(w http.ResponseWriter, r *http.Request) {
	if a.cfg.SourceManager == nil {
		http.Error(w, "source manager not configured", http.StatusServiceUnavailable)
		return
	}
	if err := a.cfg.SourceManager.RefreshEPG(context.Background()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (a *API) handleEPGPrograms(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "channel_id is required", http.StatusBadRequest)
		return
	}
	programs := a.cfg.XEPG.ProgramsFor(channelID)
	writeJSON(w, programs)
}

type channelMappingRequest struct {
	CustomName string  `json:"custom_name"`
	EPGChannel string  `json:"epg_channel"`
	ChannelNum float64 `json:"channel_num"`
}

func (a *API) handleChannelMapping(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req channelMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.CustomName != "" {
		if !a.cfg.XEPG.SetCustomName(id, req.CustomName) {
			http.Error(w, "channel not found", http.StatusNotFound)
			return
		}
	}
	if req.EPGChannel != "" {
		if !a.cfg.XEPG.SetEPGChannel(id, req.EPGChannel) {
			http.Error(w, "channel not found", http.StatusNotFound)
			return
		}
	}
	if req.ChannelNum > 0 {
		if !a.cfg.XEPG.SetChannelNum(id, req.ChannelNum) {
			http.Error(w, "channel not found", http.StatusNotFound)
			return
		}
	}
	if a.cfg.OnChannelsChanged != nil {
		a.cfg.OnChannelsChanged()
	}
	w.WriteHeader(http.StatusNoContent)
}
