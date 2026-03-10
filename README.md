# xTeVe 3

An M3U proxy server that bridges IPTV streams with **Plex DVR** and **Emby Live TV**. xTeVe presents itself as an HDHomeRun network tuner, merges M3U playlists and XMLTV EPG feeds, and re-streams channels through configurable buffer backends.

---

## Features

- **M3U management** — merge multiple playlists, filter channels, set custom names and logos
- **XMLTV / EPG** — merge external EPG sources and map them to channels
- **HDHomeRun emulation** — Plex and Emby discover xTeVe as a network tuner automatically
- **Stream buffering** — HLS pass-through, FFmpeg re-encode, or VLC transcoding
- **Tuner limits** — configure how many simultaneous streams are allowed
- **React web UI** — manage channels, playlists, settings, and live logs in the browser
- **WebSocket log stream** — real-time server logs on the Logs page
- **SSDP advertisement** — auto-discovered on the local network
- **Docker-ready** — single-binary or multi-stage Docker build

---

## Downloads

Pre-built binaries are attached to every [GitHub Release](../../releases/latest).

| Platform | Architecture | File |
|----------|-------------|------|
| Linux | x86-64 | `xteve_linux_amd64` |
| Linux | ARM64 | `xteve_linux_arm64` |
| macOS | x86-64 | `xteve_darwin_amd64` |
| macOS | Apple Silicon | `xteve_darwin_arm64` |
| Windows | x86-64 | `xteve_windows_amd64.exe` |
| FreeBSD | x86-64 | `xteve_freebsd_amd64` |

### Quick start (Linux)

```bash
curl -L https://github.com/jskoetsier/xteve2/releases/latest/download/xteve_linux_amd64 -o xteve
chmod +x xteve
./xteve
# Open http://localhost:34400
```

---

## Docker

```bash
docker run -d \
  --name xteve \
  -p 34400:34400 \
  -v ~/.xteve:/config \
  ghcr.io/jskoetsier/xteve2:latest
```

Or with docker compose:

```bash
docker compose up -d
```

---

## Build from Source

### Requirements

- [Go 1.24+](https://golang.org/dl/)
- [Node.js 22+](https://nodejs.org/) (for the web UI)

### Build

```bash
# 1. Build the React frontend
cd web && npm install && npm run build && cd ..

# 2. Build the Go binary (embeds the UI)
go build -o xteve ./cmd/xteve/

# 3. Run
./xteve -port 34400 -config ~/.xteve -debug 1
```

### Run flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `34400` | HTTP listen port |
| `-config` | `~/.xteve` | Config directory |
| `-debug` | `0` | Log verbosity (0–3) |

---

## Architecture

```
cmd/xteve/main.go        Entry point — flags, wiring, HTTP server, graceful shutdown
internal/
  config/                Settings (port, tuners, auth, FFmpeg path) — JSON persistence
  storage/               Atomic JSON file read/write with mutex
  auth/                  Optional session auth — bcrypt passwords, cookie sessions
  m3u/                   M3U parser (Parse, Filter) — produces Channel structs
  xepg/                  Channel database — Sync preserves user edits across refreshes
  hdhr/                  HDHomeRun endpoints (/discover.json, /lineup.json, /device.xml)
  ssdp/                  LAN advertisement via UPnP/SSDP (go-ssdp)
  buffer/                Tuner slot manager — enforces TunerCount limit
  api/                   REST handlers (/api/v1/*) + WebSocket hub (/ws)
  ui/                    go:embed wrapper — serves React SPA from binary
web/
  src/                   React 19 + TypeScript source (Vite, TanStack Query, shadcn/ui)
  dist/ → internal/ui/dist/   Build output embedded into the Go binary
```

### How it fits together

```
Plex / Emby
    │  discovers via SSDP + /discover.json
    │  fetches /lineup.json  →  stream URLs pointing to xTeVe
    │  fetches /xmltv/       →  merged EPG
    │
    ▼
xTeVe HTTP server (:34400)
    ├── /discover.json  /lineup.json  /device.xml   ← HDHomeRun protocol
    ├── /xmltv/  /m3u/                               ← EPG + playlist delivery
    ├── /stream/:id                                  ← stream proxy / buffer
    ├── /api/v1/*                                    ← REST API (auth-optional)
    ├── /ws                                          ← WebSocket log stream
    └── /                                            ← React SPA
```

### Data flow for a channel stream

```
Client requests /stream/:id
    → buffer.Acquire()  checks tuner limit (ErrTunerLimitReached → 503)
    → proxy upstream M3U URL via FFmpeg / VLC / HLS pass-through
    → stream bytes to client
    → buffer.Release() on disconnect
```

---

## Configuration

Config is stored in `~/.xteve/settings.json` (or the directory passed via `-config`).

| Key | Default | Description |
|-----|---------|-------------|
| `port` | `34400` | HTTP port |
| `tuner_count` | `1` | Max simultaneous streams |
| `buffer_type` | `"hls"` | `"hls"`, `"ffmpeg"`, or `"vlc"` |
| `ffmpeg_path` | `/usr/bin/ffmpeg` | Path to FFmpeg binary |
| `auth_enabled` | `false` | Enable password protection |

---

## API

All endpoints are under `/api/v1/`. Auth is enforced via session cookie when `auth_enabled = true`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/status` | Version, active streams, tuner count |
| GET | `/api/v1/settings` | Read current settings |
| PUT | `/api/v1/settings` | Update settings |
| GET | `/api/v1/channels` | List all channels with enabled state |
| PUT | `/api/v1/channels/:id` | Enable / disable a channel |
| POST | `/api/v1/auth/login` | Login (sets session cookie) |
| POST | `/api/v1/auth/logout` | Logout |

WebSocket: `ws://host:34400/ws` — streams JSON log messages `{"type":"log","msg":"..."}`.

---

## Development

```bash
# Backend (with live reload via air or just go run)
go run ./cmd/xteve/ -debug 2

# Frontend dev server (proxies /api and /ws to :34400)
cd web && npm run dev
# → http://localhost:5173
```

Run tests:

```bash
go test ./...
```

---

## Contributing

Contributions are welcome — ideas, bug reports, and code alike.

- **Have a suggestion or found a bug?** Open an [issue](../../issues) — no idea is too small.
- **Want to contribute code?** Fork the repo, make your changes, and open a pull request. AI-assisted coding is completely fine — use whatever tools help you write good code.
- **Not sure where to start?** Browse the open issues or check the architecture section above to understand the codebase.

There's no formal process. If the change makes sense, it gets merged.

---

## License

MIT — see [LICENSE](LICENSE).
