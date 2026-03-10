# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xTeVe is an M3U proxy server (Go + React) that bridges streaming services with Plex DVR and Emby Live TV. It implements the HDHomeRun protocol so media servers treat it as a network tuner.

## Build & Run

### Backend
```bash
go build -o xteve ./cmd/xteve/
./xteve                              # default port 34400, config ~/.xteve/
./xteve -port 8080 -config /data -debug 2
```

### Frontend (development)
```bash
cd web && npm install && npm run dev   # Vite on :5173, proxies /api → :34400
```

### Production build (embed UI in binary)
```bash
cd web && npm run build && cd ..
go build ./cmd/xteve/
```

### Docker
```bash
docker compose up --build
```

## Testing

```bash
go test ./...                              # all Go tests
go test ./internal/m3u/ -run TestParse    # single test
go test ./internal/... -v                  # verbose
```

## Architecture

```
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
  ui/                   go:embed wrapper for web/dist/ (build output)
web/
  src/                  React 19 + TypeScript source
  dist/                 Built output (embedded in binary via internal/ui)
```

## Key Design Decisions

- Auth middleware only wraps `/api/` — streaming endpoints are always public for Plex/Emby
- `xepg.DB.Sync()` preserves user metadata (enabled state, custom names) across playlist refreshes
- `buffer.Buffer` enforces tuner limits; returns `ErrTunerLimitReached` (→ 503) when full
- Vite build outputs to `internal/ui/dist/` so `go:embed` picks it up directly (no symlinks)
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
