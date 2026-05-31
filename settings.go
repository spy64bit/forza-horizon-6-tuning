package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const defaultListenAddr = "0.0.0.0:8000"

// Settings holds user-configurable application options persisted to disk.
type Settings struct {
	ListenAddr string `json:"listenAddr"`
}

var (
	settingsMu  sync.RWMutex
	curSettings = Settings{ListenAddr: defaultListenAddr}
)

// settingsPath returns (and lazily creates the parent dir for) the settings file.
func settingsPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	dir := filepath.Join(base, "ForzaHorizon6Telemetry")
	os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "settings.json")
}

// loadSettings reads settings from disk into curSettings, keeping defaults on error.
func loadSettings() {
	data, err := os.ReadFile(settingsPath())
	if err != nil {
		return
	}
	var s Settings
	if json.Unmarshal(data, &s) == nil && s.ListenAddr != "" {
		settingsMu.Lock()
		curSettings = s
		settingsMu.Unlock()
	}
}

// saveSettingsToDisk persists the given settings as JSON.
func saveSettingsToDisk(s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(), data, 0o644)
}

// getListenAddr returns the currently configured UDP listen address.
func getListenAddr() string {
	settingsMu.RLock()
	defer settingsMu.RUnlock()
	return curSettings.ListenAddr
}
