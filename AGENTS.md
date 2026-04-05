# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xTeVe is an M3U proxy server that bridges streaming services with Plex DVR and Emby Live TV. It merges M3U playlists and XMLTV EPG files, provides channel management, implements HDHomeRun (HDHR) protocol, and handles stream buffering via FFmpeg/VLC.

## Deployment

**Important:** For deployment to Rancher, use the `deploy-xteve` skill. See: `/Users/sebastiaan.koetsier/.agents/skills/deploy-xteve/SKILL.md`

## Build & Run

```bash
# Build all packages
go build ./...

# Run tests
go test ./...

# Run (default port 34400, config at ~/.xteve/)
./cmd/xteve/xteve

# Run with options
./cmd/xteve/xteve -port 8080 -config /path/to/config -debug 2
```

## Architecture

```
cmd/xteve/main.go       Entry point — initializes all packages
internal/
  api/                  REST handlers + WebSocket hub
  auth/                 Optional bcrypt session auth
  buffer/              Stream tuner slot management
  config/              Settings load/save (JSON)
  hdhr/                HDHomeRun discovery endpoints
  m3u/                 M3U parsing + Channel type
  source/              M3U/XMLTV source management + stream proxy
  ssdp/                LAN SSDP advertisement
  storage/             JSON file persistence
  ui/                 go:embed wrapper for React build
  xepg/              In-memory EPG database
web/                   React 19 + TypeScript frontend (Vite build)
```

## Key Data Structures

- **xepg.Entry** — channel with mapping metadata (ID, Channel, Enabled, CustomName, EPGChannel, ChannelNum)
- **xepg.Program** — programme entry (Channel, Start, Stop, Title, Desc, Category, Icon, Episode)
- **config.Settings** — user-editable config (Port, TunerCount, M3UURL, XMLTVURL, etc.)
- **source.Manager** — orchestrates M3U refresh, XMLTV proxy, HDHR lineup sync

## HTTP Endpoints

| Path | Purpose |
|------|---------|
| `/api/v1/status` | Server status |
| `/api/v1/settings` | Get/update settings |
| `/api/v1/channels` | List channels |
| `/api/v1/channels/{id}` | Update channel |
| `/api/v1/channels/{id}/mapping` | Set custom name, EPG channel, channel number |
| `/api/v1/playlists/refresh` | Refresh M3U playlist |
| `/api/v1/epg/refresh` | Refresh EPG from XMLTV |
| `/api/v1/epg/programs?channel_id=X` | Get programmes for channel |
| `/stream/{id}` | Stream proxy |
| `/xmltv/` | XMLTV EPG proxy |
| `/m3u/` | M3U playlist delivery |
| `/discover.json`, `/lineup.json`, `/device.xml` | HDHomeRun discovery |
| `/ws` | WebSocket for real-time updates |

## Configuration Storage

Default location: `~/.xteve/`

Key files: `settings.json`

## React Frontend

The web UI is in `/web/` (React 19 + TypeScript + Vite + Tailwind). The build is embedded into the Go binary via `go:embed` in `internal/ui/`.

```bash
cd web && npm install && npm run build
```

## Testing

```bash
go test ./...
```
