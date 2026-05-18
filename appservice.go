package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"lanfolder/internal/config"
	"lanfolder/internal/desktop"
	"lanfolder/internal/i18n"
	"lanfolder/internal/platform"
	"lanfolder/internal/server"
	"lanfolder/internal/share"
)

var (
	errInvalidPort              = errors.New("invalid_port")
	errSharedDirRequired        = errors.New("shared_dir_required")
	errAccessApprovalRequired   = errors.New("access_approval_required")
	errAccessRequestUnavailable = errors.New("access_request_unavailable")
)

type commandError struct {
	Code   string         `json:"error"`
	Params map[string]any `json:"params,omitempty"`
}

func newCommandError(err error, params map[string]any) error {
	return commandError{Code: err.Error(), Params: params}
}

func (e commandError) Error() string {
	return e.Code
}

func marshalCommandError(err error) []byte {
	payload := commandErrorPayload(err)
	if payload == nil {
		return nil
	}
	data, jsonErr := json.Marshal(payload)
	if jsonErr != nil {
		return nil
	}
	return data
}

func commandErrorPayload(err error) *desktop.ErrorPayload {
	var commandErr commandError
	if !errors.As(err, &commandErr) {
		return nil
	}
	return &desktop.ErrorPayload{Error: commandErr.Code, Params: commandErr.Params}
}

type AppService struct {
	mu                 sync.Mutex
	app                *application.App
	window             application.Window
	notifier           noticeNotifier
	server             *server.Server
	config             config.Config
	notices            []desktop.Notice
	noticeSeq          uint64
	drained            bool
	lastAccessNoticeAt time.Time
}

const accessNoticeCooldown = 10 * time.Second

func (s *AppService) State() desktop.AppState {
	s.mu.Lock()
	cfg := s.config
	s.mu.Unlock()
	return s.snapshot(cfg)
}

func (s *AppService) ChooseFolder() (desktop.AppState, error) {
	result, err := s.app.Dialog.OpenFile().
		CanChooseDirectories(true).
		CanCreateDirectories(true).
		CanChooseFiles(false).
		PromptForSingleSelection()
	if err != nil {
		return s.State(), err
	}
	if result == "" {
		return s.State(), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	nextConfig := s.config
	nextConfig.SharedDir = result
	if err := config.Save(nextConfig); err != nil {
		return s.snapshot(s.config), err
	}
	s.config = nextConfig
	if s.server.Status().Running {
		if err := s.restartLocked(); err != nil {
			return s.snapshot(s.config), err
		}
	}
	return s.snapshot(s.config), nil
}

func (s *AppService) SaveSettings(cfg config.Config) (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return s.snapshot(s.config), newCommandError(errInvalidPort, nil)
	}
	if !cfg.Permission.Valid() {
		cfg.Permission = share.PermissionReadOnly
	}
	cfg.Language = i18n.NormalizeLanguage(cfg.Language)
	if cfg.AutoShare && !cfg.AccessApproval {
		return s.snapshot(s.config), newCommandError(errAccessApprovalRequired, nil)
	}
	if cfg.StartAtLogin != s.config.StartAtLogin {
		if err := platform.SetStartAtLogin(cfg.StartAtLogin); err != nil {
			return s.snapshot(s.config), err
		}
	}
	wasRunning := s.server.Status().Running
	shouldRestart := wasRunning && serverConfigChanged(s.config, cfg)
	if err := config.Save(cfg); err != nil {
		return s.snapshot(s.config), err
	}
	s.config = cfg
	if shouldRestart {
		if err := s.restartLocked(); err != nil {
			return s.snapshot(s.config), err
		}
	}
	return s.snapshot(s.config), nil
}

func (s *AppService) StartSharing() (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config.SharedDir == "" {
		return s.snapshot(s.config), newCommandError(errSharedDirRequired, nil)
	}
	err := s.startSharingLocked()
	return s.snapshot(s.config), err
}

func (s *AppService) autoStartSharing() {
	s.mu.Lock()
	autoShare := s.config.AutoShare
	accessApproval := s.config.AccessApproval
	sharedDir := s.config.SharedDir
	s.mu.Unlock()
	if !autoShare {
		return
	}
	if !accessApproval {
		s.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, newCommandError(errAccessApprovalRequired, nil), "")
		return
	}
	if sharedDir == "" {
		s.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, newCommandError(errSharedDirRequired, nil), "")
		return
	}
	if _, err := s.StartSharing(); err != nil {
		s.addNotice(desktop.NoticeError, desktop.NoticeSourceStartup, err, "")
	}
}

func (s *AppService) StopSharing() (desktop.AppState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.server.Stop(ctx)
	return s.snapshot(s.config), err
}

func (s *AppService) startSharingLocked() error {
	return s.server.Start(server.Config{
		Host:           "0.0.0.0",
		Port:           s.config.Port,
		Root:           s.config.SharedDir,
		Permission:     s.config.Permission,
		ShowHidden:     s.config.ShowHiddenFiles,
		AccessApproval: s.config.AccessApproval,
		Language:       s.config.Language,
	})
}

func (s *AppService) OpenSharedFolder() error {
	s.mu.Lock()
	sharedDir := s.config.SharedDir
	s.mu.Unlock()
	if sharedDir == "" {
		return newCommandError(errSharedDirRequired, nil)
	}
	return platform.OpenFolder(sharedDir)
}

func (s *AppService) Logs() []server.LogEntry {
	return s.server.Logs()
}

func (s *AppService) PendingAccessRequests() []share.AccessRequest {
	return s.server.PendingAccessRequests()
}

func (s *AppService) AccessSessions() []share.AccessSession {
	return s.server.AccessSessions()
}

func (s *AppService) ApproveAccessRequest(id string) error {
	if err := s.server.ApproveAccessRequest(id); err != nil {
		return newCommandError(errAccessRequestUnavailable, nil)
	}
	s.emitStateChanged("access")
	return nil
}

func (s *AppService) DenyAccessRequest(id string) error {
	if err := s.server.DenyAccessRequest(id); err != nil {
		return newCommandError(errAccessRequestUnavailable, nil)
	}
	s.emitStateChanged("access")
	return nil
}

func (s *AppService) RevokeAccessSession(id string) error {
	if ok := s.server.RevokeAccessSession(id); !ok {
		return newCommandError(errAccessRequestUnavailable, nil)
	}
	s.emitStateChanged("access")
	return nil
}

func (s *AppService) handleAccessRequestNotice() {
	s.emitStateChanged("access")
	if !s.shouldShowAccessNotice(time.Now()) {
		return
	}
	s.mu.Lock()
	language := s.config.Language
	s.mu.Unlock()
	s.addNotice(desktop.NoticeInfo, desktop.NoticeSourceSystem, nil, i18n.T(language, "notice.accessRequest", nil))
}

func (s *AppService) shouldShowAccessNotice(now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.lastAccessNoticeAt.IsZero() && now.Sub(s.lastAccessNoticeAt) < accessNoticeCooldown {
		return false
	}
	s.lastAccessNoticeAt = now
	return true
}

func (s *AppService) DrainNotices() []desktop.Notice {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]desktop.Notice, len(s.notices))
	copy(out, s.notices)
	s.notices = nil
	s.drained = true
	return out
}

func (s *AppService) addNotice(level desktop.NoticeLevel, source desktop.NoticeSource, err error, message string) {
	notice := desktop.Notice{
		Level:     level,
		Source:    source,
		Message:   message,
		CreatedAt: time.Now(),
	}
	if payload := commandErrorPayload(err); payload != nil {
		notice.Error = payload
	}

	s.mu.Lock()
	s.noticeSeq++
	notice.ID = strconv.FormatUint(s.noticeSeq, 10)
	if !s.drained {
		s.notices = append(s.notices, notice)
		if len(s.notices) > 50 {
			s.notices = s.notices[len(s.notices)-50:]
		}
	}
	app := s.app
	s.mu.Unlock()

	if app != nil {
		app.Event.Emit("app:notice", notice)
	}
}

func (s *AppService) showMainWindow() {
	s.mu.Lock()
	window := s.window
	s.mu.Unlock()
	if window == nil {
		return
	}
	window.Flash(false)
	window.Show()
	window.Restore()
	window.Focus()
}

func (s *AppService) restartLocked() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.server.Stop(ctx); err != nil {
		return err
	}
	return s.startSharingLocked()
}

func (s *AppService) snapshot(cfg config.Config) desktop.AppState {
	return desktop.SnapshotBuilder{App: s.app, Window: s.window}.Build(cfg, s.server.Status(), s.addresses(cfg))
}

func (s *AppService) addresses(cfg config.Config) []string {
	ips := platform.LANIPs()
	if len(ips) == 0 {
		ips = []string{"127.0.0.1"}
	}
	addrs := make([]string, 0, len(ips))
	for _, ip := range ips {
		addrs = append(addrs, fmt.Sprintf("http://%s:%d", ip, cfg.Port))
	}
	return addrs
}

func (s *AppService) emitStateChanged(reason string) {
	if s.app != nil {
		s.app.Event.Emit("app:state-changed", map[string]string{"reason": reason})
	}
}

func serverConfigChanged(previous config.Config, next config.Config) bool {
	return previous.SharedDir != next.SharedDir ||
		previous.Port != next.Port ||
		previous.Permission != next.Permission ||
		previous.ShowHiddenFiles != next.ShowHiddenFiles ||
		previous.AccessApproval != next.AccessApproval ||
		previous.Language != next.Language
}
