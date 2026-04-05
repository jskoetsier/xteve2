# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xTeVe is an M3U proxy server that bridges streaming services with Plex DVR and Emby Live TV. It merges M3U playlists and XMLTV EPG files, provides channel management, implements HDHomeRun (HDHR) protocol, and handles stream buffering via FFmpeg/VLC.

## Build & Run

```bash
# Build
go build xteve.go

# Run (default port 34400, config at ~/.xteve/)
./xteve

# Run with options
./xteve -port 8080 -config /path/to/config -debug 2

# Development mode (serves HTML/JS from local ./html/ instead of embedded)
./xteve -dev

# Restore from backup
./xteve -restore /path/to/xteve_backup.zip
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests in a specific internal package
go test ./src/internal/m3u-parser/

# Run a single test
go test ./src/internal/m3u-parser/ -run TestName
```

Tests only exist in `src/internal/m3u-parser/`.

## TypeScript Frontend

The web UI is written in TypeScript, located in `/ts/`. Compiled output goes to `/html/js/`.

```bash
cd ts && ./compileJS.sh
```

The compiled JS is committed to the repo — edit TypeScript sources, not the compiled output.

## Architecture

```
xteve.go                  Entry point — parses flags, calls src/system.go init
src/system.go             System initialization, folder/file creation, settings bootstrap
src/webserver.go          HTTP server, all route handlers
src/xepg.go               XEPG database, channel mapping, XMLTV generation
src/m3u.go                M3U playlist parsing, filtering, channel ordering
src/buffer.go             Stream buffering/re-streaming engine (uses FFmpeg/VLC)
src/hdhr.go               HDHomeRun protocol endpoints
src/data.go               Settings persistence, provider config, runtime data updates
src/authentication.go     Auth middleware
src/backup.go             Backup/restore logic
src/internal/
  authentication/         HMAC-SHA256 token auth
  m3u-parser/             Reusable M3U parsing library (has tests)
  imgcache/               Channel logo/EPG image caching
  up2date/                GitHub-based auto-update
html/                     Web UI assets (HTML templates, CSS, compiled JS)
ts/                       TypeScript source for the web UI
```

## Key Data Structures (in `src/struct.go`)

- **SystemStruct** — global runtime state (flags, folders, version, IPs)
- **DataStruct** — runtime data (playlists, channels, filters, XEPG mapping)
- **SettingsStruct** — user-editable config (port, auth, EPG source, FFmpeg/VLC paths, tuner limits)

## HTTP Endpoints

| Path | Purpose |
|------|---------|
| `/stream/<id>` | Stream buffering/proxy |
| `/xmltv/` | XMLTV EPG delivery |
| `/m3u/` | M3U playlist delivery |
| `/api/` | JSON API (settings, channels, mapping) |
| `/web/` | Web UI |
| `/data/` | WebSocket for real-time updates |
| `/images/` | Cached channel logos |
| `/download/` | File downloads (backups, exports) |
| `/discover.json`, `/lineup.json`, `/device.xml` | HDHomeRun discovery |

## Configuration Storage

Default location: `~/.xteve/`

Key files: `settings.json`, `authentication.json`, `xepg.json`, `urls.json`, `pms.json`

Subdirectories: `backup/`, `cache/`, `temp/`, `img-cache/`, `img-upload/`
