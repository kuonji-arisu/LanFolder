package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

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

const appSingleInstanceID = "app.lanfolder.desktop"

func main() {
	staticFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	cfg := config.Load()
	if startAtLogin, err := platform.StartAtLoginEnabled(); err == nil {
		cfg.StartAtLogin = startAtLogin
	}

	notificationRuntime := &notificationRuntimeService{notifier: notifications.New()}
	appService := &AppService{
		server:   server.New(staticFS),
		config:   cfg,
		notifier: notificationRuntime,
	}
	notificationRuntime.notifier.OnNotificationResponse(func(result notifications.NotificationResult) {
		if result.Error == nil {
			appService.showMainWindow()
		}
	})
	appService.server.SetAccessRequestCallback(func() {
		appService.handleAccessRequestNotice()
	})

	app := application.New(application.Options{
		Name:           "LanFolder",
		Description:    "A minimal LAN folder sharing desktop app",
		MarshalError:   marshalCommandError,
		SingleInstance: singleInstanceOptions(appService),
		Services: []application.Service{
			application.NewService(appService),
			application.NewService(notificationRuntime),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	appService.app = app
	appService.autoStartSharing()

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
	appService.mu.Lock()
	appService.window = window
	appService.mu.Unlock()
	app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		appService.emitStateChanged("theme")
	})
	window.OnWindowEvent(events.Common.WindowFocus, func(event *application.WindowEvent) {
		window.Flash(false)
	})
	window.OnWindowEvent(events.Common.WindowShow, func(event *application.WindowEvent) {
		window.Flash(false)
	})
	window.OnWindowEvent(events.Common.WindowRestore, func(event *application.WindowEvent) {
		window.Flash(false)
	})
	setupTray(app, window, appService)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func singleInstanceOptions(appService *AppService) *application.SingleInstanceOptions {
	return &application.SingleInstanceOptions{
		UniqueID: appSingleInstanceID,
		OnSecondInstanceLaunch: func(application.SecondInstanceData) {
			appService.showMainWindow()
		},
	}
}

func setupTray(app *application.App, window application.Window, appService *AppService) {
	var forceQuit bool

	tray := app.SystemTray.New().SetIcon(appIcon)
	tray.SetTooltip("LanFolder")

	menu := app.NewMenu()
	menu.Add("显示 LanFolder").OnClick(func(*application.Context) {
		appService.showMainWindow()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(*application.Context) {
		forceQuit = true
		app.Quit()
	})
	tray.SetMenu(menu)
	tray.OnClick(func() {
		appService.showMainWindow()
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
