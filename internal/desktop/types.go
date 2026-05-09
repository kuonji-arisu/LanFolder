package desktop

import (
	"lanfolder/internal/config"
	"lanfolder/internal/server"
	"lanfolder/internal/share"
)

type AppState struct {
	Config       config.Config            `json:"config"`
	Server       server.RuntimeStatus     `json:"server"`
	AppInfo      AppInfo                  `json:"appInfo"`
	Capabilities Capabilities             `json:"capabilities"`
	Addresses    []string                 `json:"addresses"`
	Permissions  []share.PermissionOption `json:"permissions"`
}

type AppInfo struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	OS          string     `json:"os"`
	OSName      string     `json:"osName"`
	OSVersion   string     `json:"osVersion"`
	Arch        string     `json:"arch"`
	Locale      string     `json:"locale"`
	SystemTheme string     `json:"systemTheme"`
	ConfigPath  string     `json:"configPath"`
	Window      WindowInfo `json:"window"`
}

type Capabilities struct {
	StartAtLogin      bool `json:"startAtLogin"`
	Tray              bool `json:"tray"`
	OpenFolder        bool `json:"openFolder"`
	SystemThemeEvents bool `json:"systemThemeEvents"`
	WindowState       bool `json:"windowState"`
}

type WindowInfo struct {
	X          int  `json:"x"`
	Y          int  `json:"y"`
	Width      int  `json:"width"`
	Height     int  `json:"height"`
	Maximised  bool `json:"maximised"`
	Minimised  bool `json:"minimised"`
	Fullscreen bool `json:"fullscreen"`
}
