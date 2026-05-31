# Forza Horizon 6 Telemetry

A desktop app that captures, records, and replays **Forza Horizon** UDP telemetry with a live dashboard, 2D route map, and car tuning database.

Built with **Wails v2** (Go backend + Vue 3 frontend), targeting Windows.

---

## Features

- **Live telemetry** — reads UDP packets from Forza's Data Out feature and displays speed, gear, RPM, throttle, brake, boost, and more in real time
- **Session recording & replay** — record a drive to a `.bin` file and replay it with pause, resume, and seek
- **2D route map** — canvas-based route trail rendered live as you drive
- **Car tuning database** — store and manage tuning setups (tire pressure, gearing, alignment, suspension, aero, brakes, differential) backed by SQLite; supports JSON import
- **Tune selector** — pick a saved tune while in the telemetry view to display its values alongside live data

---

## Prerequisites

| Tool | Version |
|---|---|
| [Go](https://go.dev/dl/) | 1.23+ |
| [Wails CLI](https://wails.io/docs/gettingstarted/installation) | v2.12+ |
| [Node.js](https://nodejs.org/) | 18+ (LTS) |
| [npm](https://www.npmjs.com/) | 9+ |

Install Wails CLI after Go is set up:

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify everything is in order:

```bash
wails doctor
```

---

## Development

### 1. Clone and install dependencies

```bash
git clone <repo-url>
cd forza-horizon-6-tuning
```

Wails installs frontend dependencies automatically on first `wails dev`, but you can also do it manually:

```bash
cd frontend
npm install
cd ..
```

### 2. Run in dev mode (hot reload)

```bash
wails dev
```

This starts:
- The Go backend bound to the Wails WebView2 window
- A Vite dev server with hot module reload for the Vue frontend

The frontend is also accessible at `http://localhost:34115` from a browser — useful for calling Go methods from DevTools.

> **Tip:** After adding or changing a Go method on `App`, update `frontend/wailsjs/go/main/App.js` and `App.d.ts` manually so the frontend works before the next full rebuild. Running `wails dev` or `wails build` regenerates them automatically.

### 3. Configure Forza

In Forza Horizon, enable **HUD and Gameplay → Data Out**:
- Set the IP to your PC's address
- Set the port to `8000` (default) or whatever is configured in the app's Settings
- Packet format: **Car Dash** (324-byte Forza Horizon layout)

### 4. Go-only type check (fast)

```bash
go build ./...
```

```bash
go vet ./...
```

---

## Production Build

```bash
wails build
```

Output binary is written to `build/bin/`.

The NSIS installer script in `build/windows/installer/` can be used to produce a `.exe` installer.

---

## Project Structure

```
app.go          — App struct, mode state machine (live/recording/replaying), exposed JS methods
telemetry.go    — ForzaHorizonPacket (324-byte layout), packet parsing, TelemetrySnapshot
record.go       — Live listener and session recorder (UDP → .bin files)
replay.go       — Session replay with realtime pacing, pause, seek
settings.go     — JSON settings (listen address), persisted to UserConfigDir
tunes.go        — SQLite car tuning database (CRUD, import, search)
main.go         — Wails app entry point

frontend/src/
  App.vue                      — Main view: telemetry dashboard, route map, controls
  style.css                    — Global styles
  tuneSchema.js                — Shared field definitions for the tuning form
  components/
    RouteMap.vue               — Canvas 2D route trail (requestAnimationFrame loop)
    SessionCombobox.vue        — Session file picker
    TuneSelect.vue             — Tune picker combobox (telemetry view)
    TuningPanel.vue            — Car tuning CRUD modal
    TuningPanel.css            — Styles scoped to TuningPanel
```

### Data storage

| Data | Location |
|---|---|
| Session recordings (`.bin`) | `%LocalAppData%\ForzaHorizon6Telemetry\sessions\` |
| Car tuning database (`tunes.db`) | `%AppData%\ForzaHorizon6Telemetry\tunes.db` |
| Settings (`settings.json`) | `%AppData%\ForzaHorizon6Telemetry\settings.json` |

---

## Configuration

Edit `wails.json` for project metadata. See [Wails project config docs](https://wails.io/docs/reference/project-config) for all options.
