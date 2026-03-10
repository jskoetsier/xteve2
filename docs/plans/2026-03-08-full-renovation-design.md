# xTeVe Full Renovation Design

**Date:** 2026-03-08
**Approach:** Incremental migration — port logic package by package, keeping the app functional at each step. Build React frontend in parallel.

---

## Goals

- Modernize Go backend to current idioms and project layout
- Replace the 728KB compiled TypeScript blob (`webUI.go`) with a proper React app
- Drop the built-in auto-updater
- Keep all other features: HDHR, SSDP, stream buffering, M3U/XMLTV management
- Support both single-binary and Docker distribution

---

## Backend Project Layout

```
cmd/
  xteve/
    main.go              — flags, wiring, server start

internal/
  config/               — settings load/save (JSON), defaults
  auth/                 — optional bcrypt password, session middleware
  m3u/                  — playlist fetching, parsing, filtering
  xepg/                 — XMLTV parsing, EPG mapping, channel DB
  hdhr/                 — HDHomeRun discovery endpoints
  ssdp/                 — SSDP advertisement
  buffer/               — stream proxy/buffering (FFmpeg/VLC/HLS)
  api/                  — HTTP handlers (REST + WebSocket)
  storage/              — JSON file persistence (replaces scattered file I/O)

web/                    — React app source
  src/
  dist/                 — built output, embedded via go:embed

Dockerfile
docker-compose.yml
```

**Key changes from current:**
- Each `internal/` package exposes a clean interface — `api/` depends on interfaces, not concrete types
- `storage/` centralizes all file I/O (currently scattered across `system.go`, `data.go`, `xepg.go`)
- `webUI.go` (728KB blob) replaced by `go:embed web/dist`
- `up2date/` package deleted entirely
- Go bumped to 1.24

---

## API Design

### REST — `/api/v1/`

```
GET/PUT   /api/v1/settings          — server config
GET       /api/v1/status            — health, tuner counts, version

GET/POST  /api/v1/playlists         — M3U provider list
DELETE    /api/v1/playlists/:id

GET/POST  /api/v1/epg               — XMLTV provider list
DELETE    /api/v1/epg/:id

GET       /api/v1/channels          — merged channel list
PUT       /api/v1/channels/:id      — update mapping/visibility

GET       /api/v1/xepg              — XEPG mapped channel DB

POST      /api/v1/auth/login
POST      /api/v1/auth/logout
```

### Streaming (paths unchanged — Plex/Emby hardcode these)

```
GET  /stream/:id
GET  /m3u/
GET  /xmltv/
```

### HDHomeRun (paths unchanged)

```
GET  /discover.json
GET  /lineup.json
GET  /device.xml
```

### WebSocket

```
GET  /ws    — real-time log streaming and UI state push
```

---

## Frontend Architecture

**Stack:**
- React 19 + TypeScript
- Vite — build tooling, dev server with proxy to Go backend
- TanStack Query — server state, cache invalidation, background refetching
- React Router v7 — client-side routing
- shadcn/ui + Tailwind CSS — component library
- Zustand — minimal client state (auth session, UI preferences)

**Pages:**
```
/              — Dashboard (status, tuner usage, active streams)
/playlists     — M3U provider management
/epg           — XMLTV provider management
/channels      — Channel list, mapping, enable/disable, reorder
/settings      — Server config (port, auth, FFmpeg path, tuner limits)
/logs          — Live log stream via WebSocket
```

**Dev workflow:**
```bash
cd web && npm run dev    # Vite dev server :5173, proxies /api → :34400
go run ./cmd/xteve       # Backend on :34400
```

**Production:**
```bash
cd web && npm run build  # outputs to web/dist/
go build ./cmd/xteve     # embeds web/dist/ via go:embed
```

---

## Data Flow

**Playlist refresh cycle:**
```
config/providers → m3u.Fetch() → m3u.Parse() → m3u.Filter()
                                                    ↓
                              xepg.Map(channels, xmltv) → storage.Save()
                                                    ↓
                              /m3u/ and /xmltv/ serve generated output
```

Refresh runs on a configurable schedule. Also triggerable via API.

**Stream request flow:**
```
GET /stream/:id
  → xepg.LookupChannel(id)
  → buffer.AcquireTuner()     — enforces tuner limit, returns 503 if at cap
  → buffer.Proxy(streamURL)   — FFmpeg/VLC subprocess or native HLS
  → stream to client
  → buffer.ReleaseTuner() on disconnect
```

**Persistence:**
All state is JSON files in the config directory. `storage/` package owns all reads/writes with a file-lock to prevent concurrent corruption. No database.

**SSDP + HDHR:**
`ssdp` advertises on LAN at startup. `hdhr` serves static discovery endpoints reading tuner count from `config`. No behavior change, just clean package boundaries.

---

## Auth

- Config has `auth_enabled` bool (default: `false`)
- When enabled: single admin user, bcrypt-hashed password in `settings.json`
- Sessions via signed HTTP cookie
- Auth middleware wraps `/api/v1/` and `/web/` routes only
- Streaming endpoints (`/stream/`, `/m3u/`, `/xmltv/`, HDHR) always public
- First-run: if auth enabled but no password set, UI prompts for setup

---

## Distribution

**Single binary:**
```bash
go build -o xteve ./cmd/xteve
./xteve -config ~/.xteve -port 34400
```

**Docker (multi-stage):**
```
FROM node:22        → build React app
FROM golang:1.24    → build Go binary (with embedded web/dist)
FROM alpine:3.21    → final runtime image (~15MB)
```

```yaml
# docker-compose.yml
services:
  xteve:
    image: xteve
    ports: ["34400:34400"]
    volumes: ["./config:/config"]
```
