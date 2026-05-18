package main

import (
	"embed"
	"io/fs"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"lanfolder/internal/appservice"
	"lanfolder/internal/config"
	"lanfolder/internal/i18n"
	"lanfolder/internal/notice"
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

	nativeNotifier := notifications.New()
	notificationRuntime := notice.NewRuntimeService(nativeNotifier)
	shareServer := server.New(staticFS)
	appService := appservice.New(appservice.Options{
		Server:   shareServer,
		Config:   cfg,
		Notifier: notificationRuntime,
	})
	nativeNotifier.OnNotificationResponse(func(result notifications.NotificationResult) {
		if result.Error == nil {
			appservice.ShowMainWindow(appService)
		}
	})
	shareServer.SetAccessRequestCallback(func() {
		appservice.HandleAccessRequestNotice(appService)
	})

	app := application.New(application.Options{
		Name:           "LanFolder",
		Description:    "A minimal LAN folder sharing desktop app",
		MarshalError:   appservice.MarshalCommandError,
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
	appservice.SetApp(appService, app)
	appservice.AutoStartSharing(appService)
	appservice.StartAddressWatcher(appService)

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:         "LanFolder",
		Width:         350,
		Height:        580,
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
	appservice.SetWindow(appService, window)
	app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		appservice.EmitStateChanged(appService, "theme")
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

func singleInstanceOptions(appService *appservice.AppService) *application.SingleInstanceOptions {
	return &application.SingleInstanceOptions{
		UniqueID: appSingleInstanceID,
		OnSecondInstanceLaunch: func(application.SecondInstanceData) {
			appservice.ShowMainWindow(appService)
		},
	}
}

func setupTray(app *application.App, window application.Window, appService *appservice.AppService) {
	var forceQuit bool

	tray := app.SystemTray.New().SetIcon(appIcon)
	tray.SetTooltip("LanFolder")

	menu := app.NewMenu()
	language := appservice.Language(appService)
	menu.Add(i18n.T(language, "tray.show", nil)).OnClick(func(*application.Context) {
		appservice.ShowMainWindow(appService)
	})
	menu.AddSeparator()
	menu.Add(i18n.T(language, "tray.quit", nil)).OnClick(func(*application.Context) {
		forceQuit = true
		app.Quit()
	})
	tray.SetMenu(menu)
	tray.OnClick(func() {
		appservice.ShowMainWindow(appService)
	})

	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		keepInTray := appservice.KeepInTray(appService)
		if forceQuit || !keepInTray {
			return
		}
		window.Hide()
		event.Cancel()
	})
}
