package appservice

import (
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/config"
	"lanfolder/internal/desktop"
	"lanfolder/internal/notice"
	"lanfolder/internal/server"
)

type AppService struct {
	mu     sync.Mutex
	app    *application.App
	window application.Window
	config config.Config
	server *server.Server
	lanIPs func() []string

	settings   SettingsService
	sharing    *ShareService
	access     *AccessService
	notices    *NoticeCenter
	addressSvc *AddressService
	windowCtl  WindowController
}

type Options struct {
	Server   *server.Server
	Config   config.Config
	Notifier *notice.RuntimeService
	LANIPs   func() []string
}

func New(options Options) *AppService {
	return &AppService{
		server:     options.Server,
		config:     options.Config,
		lanIPs:     options.LANIPs,
		sharing:    NewShareService(options.Server),
		access:     NewAccessService(options.Server),
		notices:    NewNoticeCenter(options.Notifier),
		addressSvc: NewAddressService(options.LANIPs),
	}
}

func SetApp(service *AppService, app *application.App) {
	service.mu.Lock()
	service.app = app
	notices := service.noticeCenterLocked()
	service.mu.Unlock()
	notices.SetApp(app)
}

func SetWindow(service *AppService, window application.Window) {
	service.mu.Lock()
	service.window = window
	service.mu.Unlock()
}

func ShowMainWindow(service *AppService) {
	service.showMainWindow()
}

func AutoStartSharing(service *AppService) {
	service.autoStartSharing()
}

func StartAddressWatcher(service *AppService) {
	service.startAddressWatcher()
}

func HandleAccessRequestNotice(service *AppService) {
	service.handleAccessRequestNotice()
}

func EmitStateChanged(service *AppService, reason string) {
	service.emitStateChanged(reason)
}

func Language(service *AppService) string {
	service.mu.Lock()
	defer service.mu.Unlock()
	return service.config.Language
}

func KeepInTray(service *AppService) bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	return service.config.KeepInTray
}

func (s *AppService) State() desktop.AppState {
	s.mu.Lock()
	cfg := s.config
	s.mu.Unlock()
	return s.snapshot(cfg)
}

func (s *AppService) Logs() []server.LogEntry {
	return s.server.Logs()
}

func (s *AppService) snapshot(cfg config.Config) desktop.AppState {
	return desktop.SnapshotBuilder{App: s.app, Window: s.window}.Build(cfg, s.server.Status(), s.addressesForConfig(cfg))
}

func (s *AppService) emitStateChanged(reason string) {
	if s.app != nil {
		s.app.Event.Emit("app:state-changed", map[string]string{"reason": reason})
	}
}

func (s *AppService) noticeCenterLocked() *NoticeCenter {
	if s.notices == nil {
		s.notices = NewNoticeCenter(nil)
	}
	return s.notices
}

func (s *AppService) noticeCenter() *NoticeCenter {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.noticeCenterLocked()
}

func (s *AppService) accessServiceLocked() *AccessService {
	if s.access == nil {
		s.access = NewAccessService(s.server)
	}
	return s.access
}

func (s *AppService) accessService() *AccessService {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accessServiceLocked()
}

func (s *AppService) shareService() *ShareService {
	if s.sharing != nil {
		return s.sharing
	}
	return NewShareService(s.server)
}

func (s *AppService) addressServiceLocked() *AddressService {
	if s.addressSvc == nil {
		s.addressSvc = NewAddressService(s.lanIPs)
	}
	return s.addressSvc
}

func (s *AppService) addressService() *AddressService {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addressServiceLocked()
}
