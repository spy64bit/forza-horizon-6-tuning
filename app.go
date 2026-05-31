package main

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	mode        string // "live" | "recording" | "replaying"
	replayCmdCh chan replayCmd
}

// sessionsDir returns (and lazily creates) the directory used to store session files.
func sessionsDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = "."
	}
	dir := filepath.Join(base, "ForzaHorizon6Telemetry", "sessions")
	os.MkdirAll(dir, 0o755)
	return dir
}

// resolveSession returns the full path for a session filename.
// If the caller already passed an absolute path it is returned as-is.
func resolveSession(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(sessionsDir(), name)
}

func NewApp() *App {
	return &App{mode: "live"}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	loadSettings()
	if err := initTunesDB(); err != nil {
		wailsruntime.LogError(ctx, "tunes db init failed: "+err.Error())
	}
	// Auto-start live listener immediately
	liveCtx, cancel := context.WithCancel(ctx)
	a.mu.Lock()
	a.cancel = cancel
	a.mode = "live"
	a.mu.Unlock()
	go modeLive(liveCtx, func(p ForzaHorizonPacket) {
		wailsruntime.EventsEmit(a.ctx, "telemetry", packetToSnapshot(p))
	})
}

// resumeLive restarts the live listener in the calling goroutine.
// Must NOT be called with a.mu held.
func (a *App) resumeLive() {
	select {
	case <-a.ctx.Done():
		return
	default:
	}
	a.mu.Lock()
	liveCtx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mode = "live"
	a.mu.Unlock()
	wailsruntime.EventsEmit(a.ctx, "session:mode", "live")
	// Retry binding the UDP port — the previous goroutine may still hold it
	// for up to one read-deadline cycle (200 ms).
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}
		modeLive(liveCtx, func(p ForzaHorizonPacket) {
			wailsruntime.EventsEmit(a.ctx, "telemetry", packetToSnapshot(p))
		})
		// modeLive returns either because ctx was cancelled (stop) or because
		// the socket could not be opened.  Only loop if the context is still live.
		select {
		case <-liveCtx.Done():
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (a *App) GetSessionsDir() string {
	return sessionsDir()
}

func (a *App) ListSessions() []string {
	dir := sessionsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []string{}
	}
	var names []string
	for i := len(entries) - 1; i >= 0; i-- { // newest first
		e := entries[i]
		if !e.IsDir() && filepath.Ext(e.Name()) == ".bin" {
			names = append(names, e.Name())
		}
	}
	return names
}

func (a *App) OpenSessionsDir() {
	exec.Command("explorer", sessionsDir()).Start()
}

func (a *App) RenameSession(oldName, newName string) string {
	oldPath := resolveSession(oldName)
	newPath := resolveSession(newName)
	if err := os.Rename(oldPath, newPath); err != nil {
		return err.Error()
	}
	return ""
}

func (a *App) StartRecord(filename string) string {
	a.mu.Lock()
	if a.mode == "recording" || a.mode == "replaying" {
		a.mu.Unlock()
		return "already " + a.mode
	}
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mode = "recording"
	a.mu.Unlock()
	wailsruntime.EventsEmit(a.ctx, "session:mode", "recording")
	go func() {
		modeRecord(ctx, resolveSession(filename), func(p ForzaHorizonPacket) {
			wailsruntime.EventsEmit(a.ctx, "telemetry", packetToSnapshot(p))
		})
		a.resumeLive()
	}()
	return ""
}

func (a *App) StartReplay(filename string, realtime bool) string {
	a.mu.Lock()
	if a.mode == "recording" || a.mode == "replaying" {
		a.mu.Unlock()
		return "already " + a.mode
	}
	if a.cancel != nil {
		a.cancel()
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.mode = "replaying"
	cmdCh := make(chan replayCmd, 4)
	a.replayCmdCh = cmdCh
	a.mu.Unlock()
	wailsruntime.EventsEmit(a.ctx, "session:mode", "replaying")
	go func() {
		modeReplay(ctx, resolveSession(filename), realtime, cmdCh,
			func(p ForzaHorizonPacket) {
				wailsruntime.EventsEmit(a.ctx, "telemetry", packetToSnapshot(p))
			},
			func(frame, total int) {
				wailsruntime.EventsEmit(a.ctx, "replay:progress", map[string]int{
					"frame": frame,
					"total": total,
				})
			},
		)
		a.mu.Lock()
		a.replayCmdCh = nil
		a.mu.Unlock()
		a.resumeLive()
	}()
	return ""
}

func (a *App) StopSession() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
	a.replayCmdCh = nil
	// The active goroutine detects cancellation and calls resumeLive
}

func (a *App) sendReplayCmd(cmd replayCmd) {
	a.mu.Lock()
	ch := a.replayCmdCh
	a.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- cmd:
	default:
	}
}

func (a *App) ReplayPause() {
	a.sendReplayCmd(replayCmd{seek: -1, setPause: true, pause: true})
}

func (a *App) ReplayResume() {
	a.sendReplayCmd(replayCmd{seek: -1, setPause: true, pause: false})
}

func (a *App) ReplaySeek(frame int) {
	a.sendReplayCmd(replayCmd{seek: frame, setPause: false})
}

func (a *App) GetMode() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.mode
}

// GetListenAddr returns the currently configured UDP listen address.
func (a *App) GetListenAddr() string {
	return getListenAddr()
}

// SaveListenAddr validates and persists a new UDP listen address. It returns an
// error message on failure or "" on success. If the app is currently live, the
// listener is restarted so it rebinds to the new address immediately.
func (a *App) SaveListenAddr(addr string) string {
	if _, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return "invalid address: " + err.Error()
	}

	settingsMu.Lock()
	curSettings.ListenAddr = addr
	snapshot := curSettings
	settingsMu.Unlock()

	if err := saveSettingsToDisk(snapshot); err != nil {
		return err.Error()
	}

	// Rebind the live listener so the change takes effect now. Recording/replay
	// sessions pick up the new address when they next return to live.
	a.mu.Lock()
	mode := a.mode
	cancel := a.cancel
	a.mu.Unlock()
	if mode == "live" {
		if cancel != nil {
			cancel()
		}
		go a.resumeLive()
	}
	return ""
}
