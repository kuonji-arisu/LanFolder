package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"lanfolder/internal/config"
	"lanfolder/internal/platform"
	"lanfolder/internal/server"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var appIcon []byte

func main() {
	staticFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Load()
	if startAtLogin, err := platform.StartAtLoginEnabled(); err == nil {
		cfg.StartAtLogin = startAtLogin
	}

	appService := &AppService{
		server: server.New(staticFS),
		config: cfg,
	}

	app := application.New(application.Options{
		Name:         "LanFolder",
		Description:  "A minimal LAN folder sharing desktop app",
		MarshalError: marshalCommandError,
		Services: []application.Service{
			application.NewService(appService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	appService.app = app
	if appService.config.AutoShare && appService.config.SharedDir != "" {
		_, _ = appService.StartSharing()
	}

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:         "LanFolder",
		Width:         350,
		Height:        600,
		DisableResize: true,
		Frameless:     true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(246, 247, 250),
		URL:              "/",
	})
	appService.window = window
	app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		appService.emitStateChanged("theme")
	})
	setupTray(app, window, appService)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func setupTray(app *application.App, window application.Window, appService *AppService) {
	var forceQuit bool

	tray := app.SystemTray.New().SetIcon(appIcon)
	tray.SetTooltip("LanFolder")

	menu := app.NewMenu()
	menu.Add("显示 LanFolder").OnClick(func(*application.Context) {
		window.Show()
		window.Focus()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(*application.Context) {
		forceQuit = true
		app.Quit()
	})
	tray.SetMenu(menu)
	tray.OnClick(func() {
		window.Show()
		window.Focus()
	})

	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		appService.mu.Lock()
		keepInTray := appService.config.KeepInTray
		appService.mu.Unlock()
		if forceQuit || !keepInTray {
			return
		}
		window.Hide()
		event.Cancel()
	})
}
