package desktop

import (
	"runtime"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/buildinfo"
	"lanfolder/internal/config"
	"lanfolder/internal/platform"
	"lanfolder/internal/server"
	"lanfolder/internal/share"
)

type SnapshotBuilder struct {
	App    *application.App
	Window application.Window
}

func (b SnapshotBuilder) Build(cfg config.Config, status server.RuntimeStatus, addresses []string) AppState {
	if enabled, err := platform.StartAtLoginEnabled(); err == nil {
		cfg.StartAtLogin = enabled
	}
	return AppState{
		Config:       cfg,
		Server:       status,
		AppInfo:      b.appInfo(),
		Capabilities: b.capabilities(),
		Addresses:    addresses,
		Permissions:  share.PermissionOptions(cfg.Language),
	}
}

func (b SnapshotBuilder) appInfo() AppInfo {
	theme := "light"
	osName := runtime.GOOS
	osVersion := ""
	osValue := runtime.GOOS
	arch := runtime.GOARCH
	if b.App != nil {
		info := b.App.Env.Info()
		osValue = info.OS
		arch = info.Arch
		if b.App.Env.IsDarkMode() {
			theme = "dark"
		}
		if info.OSInfo != nil {
			if info.OSInfo.Name != "" {
				osName = info.OSInfo.Name
			}
			osVersion = info.OSInfo.Version
		}
	}
	cfgPath, _ := config.Path()
	return AppInfo{
		Name:        "LanFolder",
		Version:     buildinfo.Version,
		OS:          osValue,
		OSName:      osName,
		OSVersion:   osVersion,
		Arch:        arch,
		Locale:      platform.Locale(),
		SystemTheme: theme,
		ConfigPath:  cfgPath,
		Window:      b.windowInfo(),
	}
}

func (b SnapshotBuilder) capabilities() Capabilities {
	return Capabilities{
		StartAtLogin:      runtime.GOOS == "windows",
		Tray:              true,
		OpenFolder:        true,
		SystemThemeEvents: runtime.GOOS == "windows" || runtime.GOOS == "darwin" || runtime.GOOS == "linux",
		WindowState:       b.Window != nil,
	}
}

func (b SnapshotBuilder) windowInfo() WindowInfo {
	if b.Window == nil {
		return WindowInfo{}
	}
	x, y := b.Window.Position()
	width, height := b.Window.Size()
	return WindowInfo{
		X:          x,
		Y:          y,
		Width:      width,
		Height:     height,
		Maximised:  b.Window.IsMaximised(),
		Minimised:  b.Window.IsMinimised(),
		Fullscreen: b.Window.IsFullscreen(),
	}
}
