package appservice

import "github.com/wailsapp/wails/v3/pkg/application"

type WindowController struct{}

func (WindowController) ShowMainWindow(window application.Window) {
	if window == nil {
		return
	}
	window.Flash(false)
	window.Show()
	window.Restore()
	window.Focus()
}

func (s *AppService) showMainWindow() {
	s.mu.Lock()
	window := s.window
	controller := s.windowCtl
	s.mu.Unlock()
	controller.ShowMainWindow(window)
}
